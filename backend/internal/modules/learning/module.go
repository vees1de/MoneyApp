package learning

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"moneyapp/backend/internal/modules/catalog"
	"moneyapp/backend/internal/modules/org"
	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/clock"
	"moneyapp/backend/internal/platform/db"
	"moneyapp/backend/internal/platform/events"
	"moneyapp/backend/internal/platform/httpx"
	"moneyapp/backend/internal/platform/outbox"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Assignment struct {
	ID             uuid.UUID  `json:"id"`
	CourseID       uuid.UUID  `json:"course_id"`
	AssignmentType string     `json:"assignment_type"`
	TargetType     string     `json:"target_type"`
	TargetID       uuid.UUID  `json:"target_id"`
	AssignedBy     uuid.UUID  `json:"assigned_by"`
	Priority       string     `json:"priority"`
	Reason         *string    `json:"reason,omitempty"`
	StartAt        *time.Time `json:"start_at,omitempty"`
	DeadlineAt     *time.Time `json:"deadline_at,omitempty"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type Enrollment struct {
	ID                uuid.UUID  `json:"id"`
	CourseID          uuid.UUID  `json:"course_id"`
	UserID            uuid.UUID  `json:"user_id"`
	AssignmentID      *uuid.UUID `json:"assignment_id,omitempty"`
	Source            string     `json:"source"`
	Status            string     `json:"status"`
	EnrolledAt        time.Time  `json:"enrolled_at"`
	StartedAt         *time.Time `json:"started_at,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	DeadlineAt        *time.Time `json:"deadline_at,omitempty"`
	LastActivityAt    *time.Time `json:"last_activity_at,omitempty"`
	CompletionPercent string     `json:"completion_percent"`
	IsMandatory       bool       `json:"is_mandatory"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type CreateAssignmentRequest struct {
	CourseID       uuid.UUID  `json:"course_id" validate:"required"`
	AssignmentType string     `json:"assignment_type" validate:"required,oneof=individual department group role_based"`
	TargetType     string     `json:"target_type" validate:"required,oneof=user department group"`
	TargetID       uuid.UUID  `json:"target_id" validate:"required"`
	Priority       string     `json:"priority" validate:"required,oneof=mandatory recommended"`
	Reason         *string    `json:"reason,omitempty"`
	StartAt        *time.Time `json:"start_at,omitempty"`
	DeadlineAt     *time.Time `json:"deadline_at,omitempty"`
}

type ProgressRequest struct {
	CourseModuleID  uuid.UUID  `json:"course_module_id" validate:"required"`
	Status          string     `json:"status" validate:"required,oneof=not_started in_progress completed"`
	ProgressPercent string     `json:"progress_percent" validate:"required"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

type CompleteRequest struct {
	CompletionType string  `json:"completion_type" validate:"required,oneof=auto manual certificate_verified trainer_confirmed"`
	Score          *string `json:"score,omitempty"`
	Notes          *string `json:"notes,omitempty"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(database *sql.DB) *Repository {
	return &Repository{db: database}
}

func (r *Repository) base(exec ...db.DBTX) db.DBTX {
	if len(exec) > 0 && exec[0] != nil {
		return exec[0]
	}
	return r.db
}

func (r *Repository) CreateAssignment(ctx context.Context, item Assignment, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into course_assignments (
			id, course_id, assignment_type, target_type, target_id, assigned_by, priority,
			reason, start_at, deadline_at, status, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, item.ID, item.CourseID, item.AssignmentType, item.TargetType, item.TargetID, item.AssignedBy, item.Priority,
		item.Reason, item.StartAt, item.DeadlineAt, item.Status, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repository) ListAssignments(ctx context.Context, exec ...db.DBTX) ([]Assignment, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select id, course_id, assignment_type, target_type, target_id, assigned_by, priority,
		       reason, start_at, deadline_at, status, created_at, updated_at
		from course_assignments
		order by created_at desc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Assignment
	for rows.Next() {
		var item Assignment
		if err := rows.Scan(&item.ID, &item.CourseID, &item.AssignmentType, &item.TargetType, &item.TargetID,
			&item.AssignedBy, &item.Priority, &item.Reason, &item.StartAt, &item.DeadlineAt, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateEnrollment(ctx context.Context, item Enrollment, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into enrollments (
			id, course_id, user_id, assignment_id, source, status, enrolled_at, started_at, completed_at,
			deadline_at, last_activity_at, completion_percent, is_mandatory, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, nullif($12, '')::numeric, $13, $14, $15)
		on conflict (course_id, user_id, source, assignment_id) do nothing
	`, item.ID, item.CourseID, item.UserID, item.AssignmentID, item.Source, item.Status, item.EnrolledAt, item.StartedAt,
		item.CompletedAt, item.DeadlineAt, item.LastActivityAt, item.CompletionPercent, item.IsMandatory, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repository) GetEnrollment(ctx context.Context, id uuid.UUID, exec ...db.DBTX) (Enrollment, error) {
	var item Enrollment
	err := r.base(exec...).QueryRowContext(ctx, `
		select id, course_id, user_id, assignment_id, source, status, enrolled_at, started_at, completed_at,
		       deadline_at, last_activity_at, completion_percent::text, is_mandatory, created_at, updated_at
		from enrollments
		where id = $1
	`, id).Scan(&item.ID, &item.CourseID, &item.UserID, &item.AssignmentID, &item.Source, &item.Status, &item.EnrolledAt,
		&item.StartedAt, &item.CompletedAt, &item.DeadlineAt, &item.LastActivityAt, &item.CompletionPercent, &item.IsMandatory, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repository) ListEnrollmentsByUser(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) ([]Enrollment, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select id, course_id, user_id, assignment_id, source, status, enrolled_at, started_at, completed_at,
		       deadline_at, last_activity_at, completion_percent::text, is_mandatory, created_at, updated_at
		from enrollments
		where user_id = $1
		order by created_at desc
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Enrollment
	for rows.Next() {
		var item Enrollment
		if err := rows.Scan(&item.ID, &item.CourseID, &item.UserID, &item.AssignmentID, &item.Source, &item.Status,
			&item.EnrolledAt, &item.StartedAt, &item.CompletedAt, &item.DeadlineAt, &item.LastActivityAt,
			&item.CompletionPercent, &item.IsMandatory, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) UpdateEnrollment(ctx context.Context, item Enrollment, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update enrollments
		set status = $2,
		    started_at = $3,
		    completed_at = $4,
		    deadline_at = $5,
		    last_activity_at = $6,
		    completion_percent = nullif($7, '')::numeric,
		    is_mandatory = $8,
		    updated_at = $9
		where id = $1
	`, item.ID, item.Status, item.StartedAt, item.CompletedAt, item.DeadlineAt, item.LastActivityAt, item.CompletionPercent, item.IsMandatory, item.UpdatedAt)
	return err
}

func (r *Repository) UpsertModuleProgress(ctx context.Context, enrollmentID uuid.UUID, req ProgressRequest, updatedAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into module_progress (
			id, enrollment_id, course_module_id, status, progress_percent, started_at, completed_at, updated_at
		)
		values (
			$1, $2, $3, $4, nullif($5, '')::numeric,
			case when $4 = 'in_progress' then $6 else null end,
			$7,
			$6
		)
		on conflict (enrollment_id, course_module_id) do update
		set status = excluded.status,
		    progress_percent = excluded.progress_percent,
		    completed_at = excluded.completed_at,
		    updated_at = excluded.updated_at
	`, uuid.New(), enrollmentID, req.CourseModuleID, req.Status, req.ProgressPercent, updatedAt, req.CompletedAt)
	return err
}

func (r *Repository) CreateCompletionRecord(ctx context.Context, enrollmentID, actorID uuid.UUID, completionType string, score *string, notes *string, completedAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into completion_records (
			id, enrollment_id, completion_type, completed_by, score, completed_at, notes
		)
		values ($1, $2, $3, $4, nullif($5, '')::numeric, $6, $7)
		on conflict (enrollment_id) do update
		set completion_type = excluded.completion_type,
		    completed_by = excluded.completed_by,
		    score = excluded.score,
		    completed_at = excluded.completed_at,
		    notes = excluded.notes
	`, uuid.New(), enrollmentID, completionType, actorID, score, completedAt, notes)
	return err
}

type Service struct {
	db         *sql.DB
	repo       *Repository
	orgService *org.Service
	catalog    *catalog.Service
	outbox     *outbox.Service
	clock      clock.Clock
}

func NewService(database *sql.DB, repo *Repository, orgService *org.Service, catalogService *catalog.Service, outboxService *outbox.Service, appClock clock.Clock) *Service {
	return &Service{
		db:         database,
		repo:       repo,
		orgService: orgService,
		catalog:    catalogService,
		outbox:     outboxService,
		clock:      appClock,
	}
}

func (s *Service) CreateAssignment(ctx context.Context, principal platformauth.Principal, req CreateAssignmentRequest) (Assignment, error) {
	course, err := s.catalog.GetCourse(ctx, req.CourseID)
	if err != nil {
		return Assignment{}, err
	}

	targetUsers, err := s.orgService.ResolveTargetUsers(ctx, req.TargetType, req.TargetID)
	if err != nil {
		return Assignment{}, httpx.BadRequest("invalid_target", "assignment target could not be resolved")
	}
	if len(targetUsers) == 0 {
		return Assignment{}, httpx.BadRequest("empty_target", "assignment target resolved to zero users")
	}

	now := s.clock.Now()
	item := Assignment{
		ID:             uuid.New(),
		CourseID:       req.CourseID,
		AssignmentType: req.AssignmentType,
		TargetType:     req.TargetType,
		TargetID:       req.TargetID,
		AssignedBy:     principal.UserID,
		Priority:       req.Priority,
		Reason:         req.Reason,
		StartAt:        req.StartAt,
		DeadlineAt:     req.DeadlineAt,
		Status:         "active",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err = db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.repo.CreateAssignment(ctx, item, tx); err != nil {
			return err
		}

		for _, userID := range targetUsers {
			enrollment := Enrollment{
				ID:                uuid.New(),
				CourseID:          req.CourseID,
				UserID:            userID,
				AssignmentID:      &item.ID,
				Source:            "assignment",
				Status:            "not_started",
				EnrolledAt:        now,
				DeadlineAt:        req.DeadlineAt,
				CompletionPercent: "0",
				IsMandatory:       req.Priority == "mandatory" || course.IsMandatoryDefault,
				CreatedAt:         now,
				UpdatedAt:         now,
			}
			if err := s.repo.CreateEnrollment(ctx, enrollment, tx); err != nil {
				return err
			}

			if _, err := tx.ExecContext(ctx, `
				insert into notifications (
					id, user_id, channel, type, title, body, status, related_entity_type, related_entity_id, created_at
				)
				values ($1, $2, 'in_app', 'course_assigned', $3, $4, 'pending', 'course_assignment', $5, $6)
			`, uuid.New(), userID, "New assigned course", "You have been assigned: "+course.Title, item.ID, now); err != nil {
				return err
			}
		}

		return s.outbox.Publish(ctx, tx, events.Message{
			Topic:      "learning",
			EventType:  "learning.assignment.created",
			EntityType: "course_assignment",
			EntityID:   item.ID,
			Payload: map[string]any{
				"course_id":    item.CourseID,
				"target_type":  item.TargetType,
				"target_count": len(targetUsers),
				"assigned_by":  item.AssignedBy,
			},
			OccurredAt: now,
		})
	})

	return item, err
}

func (s *Service) ListAssignments(ctx context.Context) ([]Assignment, error) {
	return s.repo.ListAssignments(ctx)
}

func (s *Service) ListMyEnrollments(ctx context.Context, principal platformauth.Principal) ([]Enrollment, error) {
	return s.repo.ListEnrollmentsByUser(ctx, principal.UserID)
}

func (s *Service) GetEnrollment(ctx context.Context, principal platformauth.Principal, id uuid.UUID) (Enrollment, error) {
	item, err := s.repo.GetEnrollment(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Enrollment{}, httpx.NotFound("enrollment_not_found", "enrollment not found")
		}
		return Enrollment{}, err
	}
	if item.UserID != principal.UserID && !principal.HasPermission("enrollments.read") {
		return Enrollment{}, httpx.Forbidden("forbidden", "permission denied")
	}
	return item, nil
}

func (s *Service) StartEnrollment(ctx context.Context, principal platformauth.Principal, id uuid.UUID) (Enrollment, error) {
	item, err := s.GetEnrollment(ctx, principal, id)
	if err != nil {
		return Enrollment{}, err
	}
	if item.UserID != principal.UserID && !principal.HasPermission("enrollments.manage") {
		return Enrollment{}, httpx.Forbidden("forbidden", "only owner or HR/admin can start enrollment")
	}
	if item.Status == "completed" || item.Status == "canceled" {
		return Enrollment{}, httpx.BadRequest("invalid_status", "completed or canceled enrollment cannot be started")
	}
	now := s.clock.Now()
	if item.StartedAt == nil {
		item.StartedAt = &now
	}
	item.Status = "in_progress"
	item.LastActivityAt = &now
	item.UpdatedAt = now
	return item, s.repo.UpdateEnrollment(ctx, item)
}

func (s *Service) ProgressEnrollment(ctx context.Context, principal platformauth.Principal, id uuid.UUID, req ProgressRequest) (Enrollment, error) {
	item, err := s.GetEnrollment(ctx, principal, id)
	if err != nil {
		return Enrollment{}, err
	}
	if item.UserID != principal.UserID && !principal.HasPermission("enrollments.manage") {
		return Enrollment{}, httpx.Forbidden("forbidden", "only owner or HR/admin can update enrollment progress")
	}
	now := s.clock.Now()
	if err := s.repo.UpsertModuleProgress(ctx, id, req, now); err != nil {
		return Enrollment{}, err
	}
	item.Status = "in_progress"
	item.LastActivityAt = &now
	item.CompletionPercent = req.ProgressPercent
	item.UpdatedAt = now
	return item, s.repo.UpdateEnrollment(ctx, item)
}

func (s *Service) CompleteEnrollment(ctx context.Context, principal platformauth.Principal, id uuid.UUID, req CompleteRequest) (Enrollment, error) {
	item, err := s.GetEnrollment(ctx, principal, id)
	if err != nil {
		return Enrollment{}, err
	}
	if item.UserID != principal.UserID && !principal.HasPermission("enrollments.manage") {
		return Enrollment{}, httpx.Forbidden("forbidden", "only owner or HR/admin can complete enrollment")
	}
	now := s.clock.Now()
	item.Status = "completed"
	item.CompletedAt = &now
	item.LastActivityAt = &now
	item.CompletionPercent = "100"
	item.UpdatedAt = now

	err = db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.repo.UpdateEnrollment(ctx, item, tx); err != nil {
			return err
		}
		if err := s.repo.CreateCompletionRecord(ctx, item.ID, principal.UserID, req.CompletionType, req.Score, req.Notes, now, tx); err != nil {
			return err
		}
		return s.outbox.Publish(ctx, tx, events.Message{
			Topic:      "learning",
			EventType:  "learning.enrollment.completed",
			EntityType: "enrollment",
			EntityID:   item.ID,
			Payload: map[string]any{
				"user_id":   item.UserID,
				"course_id": item.CourseID,
			},
			OccurredAt: now,
		})
	})
	return item, err
}

type Handler struct {
	service   *Service
	validator *validator.Validate
}

func NewHandler(service *Service, validator *validator.Validate) *Handler {
	return &Handler{service: service, validator: validator}
}

func learningPrincipal(r *http.Request) (platformauth.Principal, error) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		return platformauth.Principal{}, httpx.Unauthorized("unauthorized", "authorization required")
	}
	return principal, nil
}

func (h *Handler) CreateAssignment(w http.ResponseWriter, r *http.Request) {
	principal, err := learningPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	var req CreateAssignmentRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}
	item, err := h.service.CreateAssignment(r.Context(), principal, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) ListAssignments(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListAssignments(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) ListMyEnrollments(w http.ResponseWriter, r *http.Request) {
	principal, err := learningPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	items, err := h.service.ListMyEnrollments(r.Context(), principal)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) GetEnrollment(w http.ResponseWriter, r *http.Request) {
	principal, err := learningPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_enrollment_id", "invalid enrollment id"))
		return
	}
	item, err := h.service.GetEnrollment(r.Context(), principal, id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) StartEnrollment(w http.ResponseWriter, r *http.Request) {
	principal, err := learningPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_enrollment_id", "invalid enrollment id"))
		return
	}
	item, err := h.service.StartEnrollment(r.Context(), principal, id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) ProgressEnrollment(w http.ResponseWriter, r *http.Request) {
	principal, err := learningPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_enrollment_id", "invalid enrollment id"))
		return
	}
	var req ProgressRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}
	item, err := h.service.ProgressEnrollment(r.Context(), principal, id, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) CompleteEnrollment(w http.ResponseWriter, r *http.Request) {
	principal, err := learningPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_enrollment_id", "invalid enrollment id"))
		return
	}
	var req CompleteRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}
	item, err := h.service.CompleteEnrollment(r.Context(), principal, id, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}
