package ai_recommendations

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"moneyapp/backend/internal/core/audit"
	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/db"
	"moneyapp/backend/internal/platform/httpx"
	platformworker "moneyapp/backend/internal/platform/worker"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const (
	aiRecommendationJobQueueName = "ai"
	aiRecommendationJobTypeRun   = "ai_recommendations_run"
)

type AIRecommendationJob struct {
	ID         uuid.UUID          `json:"id"`
	UserID     uuid.UUID          `json:"user_id"`
	Status     string             `json:"status"`
	Attempt    int                `json:"attempt"`
	Result     *RecommendResponse `json:"result,omitempty"`
	LastError  *string            `json:"last_error,omitempty"`
	StartedAt  *time.Time         `json:"started_at,omitempty"`
	FinishedAt *time.Time         `json:"finished_at,omitempty"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

type recommendationJobPayload struct {
	JobID             uuid.UUID  `json:"job_id"`
	UserID            uuid.UUID  `json:"user_id"`
	RoleCodes         []string   `json:"role_codes,omitempty"`
	PermissionCodes   []string   `json:"permission_codes,omitempty"`
	EmployeeProfileID *uuid.UUID `json:"employee_profile_id,omitempty"`
	DepartmentID      *uuid.UUID `json:"department_id,omitempty"`
	IP                *string    `json:"ip,omitempty"`
	UserAgent         *string    `json:"user_agent,omitempty"`
}

func (s *Service) StartRecommendationJob(ctx context.Context, principal platformauth.Principal, options RecommendOptions) (AIRecommendationJob, error) {
	if _, err := s.getActiveYougileConnection(ctx, principal.UserID); err != nil {
		return AIRecommendationJob{}, err
	}
	if s.queue == nil {
		return AIRecommendationJob{}, fmt.Errorf("ai recommendation queue is not configured")
	}

	now := time.Now().UTC()
	job := AIRecommendationJob{
		ID:        uuid.New(),
		UserID:    principal.UserID,
		Status:    "pending",
		Attempt:   0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return AIRecommendationJob{}, err
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.createRecommendationJob(ctx, job, tx); err != nil {
		return AIRecommendationJob{}, err
	}

	payload := recommendationJobPayload{
		JobID:             job.ID,
		UserID:            principal.UserID,
		RoleCodes:         append([]string{}, principal.RoleCodes...),
		PermissionCodes:   append([]string{}, principal.PermissionCodes...),
		EmployeeProfileID: principal.EmployeeProfileID,
		DepartmentID:      principal.DepartmentID,
		IP:                options.IP,
		UserAgent:         options.UserAgent,
	}
	idempotencyKey := "ai-recommendations:" + job.ID.String()
	if err := s.queue.Enqueue(ctx, tx, aiRecommendationJobQueueName, aiRecommendationJobTypeRun, payload, &idempotencyKey, now); err != nil {
		return AIRecommendationJob{}, err
	}
	if err := tx.Commit(); err != nil {
		return AIRecommendationJob{}, err
	}

	s.recordRecommendationJobAudit(ctx, "ai.recommendations.started", job, options, map[string]any{
		"status":  job.Status,
		"attempt": job.Attempt,
	})

	return job, nil
}

func (s *Service) GetRecommendationJob(ctx context.Context, principal platformauth.Principal, id uuid.UUID) (AIRecommendationJob, error) {
	var item AIRecommendationJob
	var resultText sql.NullString
	var lastError sql.NullString

	err := s.db.QueryRowContext(ctx, `
		select id, user_id, status, attempt, result::text, last_error, started_at, finished_at, created_at, updated_at
		from ai_recommendation_jobs
		where id = $1 and user_id = $2
	`, id, principal.UserID).Scan(
		&item.ID,
		&item.UserID,
		&item.Status,
		&item.Attempt,
		&resultText,
		&lastError,
		&item.StartedAt,
		&item.FinishedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AIRecommendationJob{}, httpx.NotFound("ai_recommendation_job_not_found", "ai recommendation job not found")
		}
		return AIRecommendationJob{}, err
	}

	if resultText.Valid && resultText.String != "" && resultText.String != "null" {
		var result RecommendResponse
		if err := json.Unmarshal([]byte(resultText.String), &result); err != nil {
			return AIRecommendationJob{}, err
		}
		item.Result = &result
	}
	if lastError.Valid {
		value := lastError.String
		item.LastError = &value
	}

	return item, nil
}

func (s *Service) ProcessRecommendationJob(ctx context.Context, job platformworker.Job) error {
	var payload recommendationJobPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return err
	}
	if payload.JobID == uuid.Nil {
		return fmt.Errorf("ai recommendation job payload is missing job_id")
	}

	now := time.Now().UTC()
	if err := s.markRecommendationJobProcessing(ctx, payload.JobID, now); err != nil {
		return err
	}

	principal := platformauth.Principal{
		UserID:            payload.UserID,
		RoleCodes:         append([]string{}, payload.RoleCodes...),
		PermissionCodes:   append([]string{}, payload.PermissionCodes...),
		EmployeeProfileID: payload.EmployeeProfileID,
		DepartmentID:      payload.DepartmentID,
	}
	principalCtx := platformauth.ContextWithPrincipal(ctx, principal)
	result, err := s.Recommend(principalCtx, principal, RecommendOptions{
		IP:        payload.IP,
		UserAgent: payload.UserAgent,
	})
	if err != nil {
		finishedAt := time.Now().UTC()
		if job.Attempt+1 >= job.MaxAttempts {
			_ = s.markRecommendationJobFailed(ctx, payload.JobID, finishedAt, err.Error())
			s.recordRecommendationJobAudit(ctx, "ai.recommendations.failed", AIRecommendationJob{
				ID:      payload.JobID,
				UserID:  payload.UserID,
				Status:  "failed",
				Attempt: job.Attempt + 1,
			}, RecommendOptions{IP: payload.IP, UserAgent: payload.UserAgent}, map[string]any{
				"error": err.Error(),
			})
		} else {
			_ = s.markRecommendationJobRetry(ctx, payload.JobID, time.Now().UTC(), err.Error())
		}
		return err
	}

	if err := s.markRecommendationJobDone(ctx, payload.JobID, result, time.Now().UTC()); err != nil {
		return err
	}

	s.recordRecommendationJobAudit(ctx, "ai.recommendations.completed_async", AIRecommendationJob{
		ID:      payload.JobID,
		UserID:  payload.UserID,
		Status:  "done",
		Attempt: job.Attempt + 1,
		Result:  &result,
	}, RecommendOptions{IP: payload.IP, UserAgent: payload.UserAgent}, map[string]any{
		"tasks_analyzed":               result.Tasks,
		"courses_in_pool":              result.CoursesInPool,
		"intakes_in_pool":              result.IntakesInPool,
		"course_recommendations_count": len(result.Recommendations),
		"intake_recommendations_count": len(result.IntakeRecommendations),
	})

	return nil
}

func (s *Service) createRecommendationJob(ctx context.Context, job AIRecommendationJob, exec ...db.DBTX) error {
	_, err := s.recommendationDB(exec...).ExecContext(ctx, `
		insert into ai_recommendation_jobs (
			id, user_id, status, attempt, result, last_error, started_at, finished_at, created_at, updated_at
		)
		values ($1, $2, $3, $4, null, null, null, null, $5, $6)
	`, job.ID, job.UserID, job.Status, job.Attempt, job.CreatedAt, job.UpdatedAt)
	return err
}

func (s *Service) recommendationDB(exec ...db.DBTX) db.DBTX {
	if len(exec) > 0 && exec[0] != nil {
		return exec[0]
	}
	return s.db
}

func (s *Service) markRecommendationJobProcessing(ctx context.Context, id uuid.UUID, startedAt time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		update ai_recommendation_jobs
		set status = 'processing',
		    started_at = coalesce(started_at, $2),
		    attempt = attempt + 1,
		    last_error = null,
		    updated_at = $2
		where id = $1
	`, id, startedAt)
	return err
}

func (s *Service) markRecommendationJobRetry(ctx context.Context, id uuid.UUID, updatedAt time.Time, message string) error {
	_, err := s.db.ExecContext(ctx, `
		update ai_recommendation_jobs
		set status = 'retry',
		    last_error = $3,
		    updated_at = $2
		where id = $1
	`, id, updatedAt, message)
	return err
}

func (s *Service) markRecommendationJobDone(ctx context.Context, id uuid.UUID, result RecommendResponse, finishedAt time.Time) error {
	payload, err := json.Marshal(result)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		update ai_recommendation_jobs
		set status = 'done',
		    result = $3::jsonb,
		    last_error = null,
		    finished_at = $2,
		    updated_at = $2
		where id = $1
	`, id, finishedAt, payload)
	return err
}

func (s *Service) markRecommendationJobFailed(ctx context.Context, id uuid.UUID, finishedAt time.Time, message string) error {
	_, err := s.db.ExecContext(ctx, `
		update ai_recommendation_jobs
		set status = 'failed',
		    last_error = $3,
		    finished_at = $2,
		    updated_at = $2
		where id = $1
	`, id, finishedAt, message)
	return err
}

func (s *Service) recordRecommendationJobAudit(ctx context.Context, action string, job AIRecommendationJob, options RecommendOptions, meta map[string]any) {
	if s.auditService == nil {
		return
	}

	if meta == nil {
		meta = map[string]any{}
	}
	meta["job_id"] = job.ID.String()
	meta["job_status"] = job.Status
	meta["job_attempt"] = job.Attempt

	actorID := job.UserID
	_ = s.auditService.RecordChange(ctx, audit.RecordInput{
		UserID:     job.UserID,
		Action:     action,
		EntityType: "ai_recommendation_jobs",
		EntityID:   &job.ID,
		Meta:       meta,
		ChangeSet: map[string]any{
			"after": map[string]any{
				"job_id":      job.ID.String(),
				"job_status":  job.Status,
				"job_attempt": job.Attempt,
			},
		},
		Source:      audit.SourceSystem,
		ActorType:   "user",
		ActorID:     &actorID,
		ActorUserID: &actorID,
		IP:          options.IP,
		UserAgent:   options.UserAgent,
	})
}

func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authentication required"))
		return
	}

	job, err := h.service.StartRecommendationJob(r.Context(), principal, RecommendOptions{
		IP:        requestIP(r),
		UserAgent: optionalString(strings.TrimSpace(r.UserAgent())),
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusAccepted, job)
}

func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authentication required"))
		return
	}

	jobID, err := uuid.Parse(chi.URLParam(r, "jobId"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_job_id", "invalid job id"))
		return
	}

	job, err := h.service.GetRecommendationJob(r.Context(), principal, jobID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, job)
}
