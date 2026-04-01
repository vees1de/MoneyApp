package analytics

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/httpx"
	"moneyapp/backend/internal/platform/worker"

	"github.com/google/uuid"
)

type Service struct {
	db    *sql.DB
	queue *worker.Queue
}

func NewService(database *sql.DB, queue *worker.Queue) *Service {
	return &Service{db: database, queue: queue}
}

func (s *Service) DashboardHR(ctx context.Context) (map[string]any, error) {
	return s.simpleCounts(ctx, map[string]string{
		"users":             "select count(*) from users",
		"courses":           "select count(*) from courses",
		"enrollments":       "select count(*) from enrollments",
		"external_requests": "select count(*) from external_course_requests",
		"pending_approvals": "select count(*) from approval_steps where status = 'pending'",
	})
}

func (s *Service) DashboardManager(ctx context.Context, principal platformauth.Principal) (map[string]any, error) {
	return s.simpleCounts(ctx, map[string]string{
		"team_requests":    "select count(*) from approval_steps where approver_user_id = '" + principal.UserID.String() + "' and status = 'pending'",
		"team_enrollments": "select count(*) from manager_relations mr join enrollments e on e.user_id = mr.employee_user_id where mr.manager_user_id = '" + principal.UserID.String() + "'",
	})
}

func (s *Service) Compliance(ctx context.Context) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `
		select department_id, mandatory_assigned_count, mandatory_completed_count, overdue_count
		from reporting_mandatory_training_compliance
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []map[string]any
	for rows.Next() {
		var departmentID *uuid.UUID
		var assigned, completed, overdue int
		if err := rows.Scan(&departmentID, &assigned, &completed, &overdue); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{
			"department_id":             departmentID,
			"mandatory_assigned_count":  assigned,
			"mandatory_completed_count": completed,
			"overdue_count":             overdue,
		})
	}
	return items, rows.Err()
}

func (s *Service) ExternalRequests(ctx context.Context) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `select status, total from reporting_external_request_funnel order by status asc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []map[string]any
	for rows.Next() {
		var status string
		var total int64
		if err := rows.Scan(&status, &total); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"status": status, "total": total})
	}
	return items, rows.Err()
}

func (s *Service) Budget(ctx context.Context) (map[string]any, error) {
	return s.simpleCounts(ctx, map[string]string{
		"limits":        "select count(*) from budget_limits",
		"reservations":  "select count(*) from budget_consumptions where status = 'reserved'",
		"spent_records": "select count(*) from budget_consumptions where status = 'spent'",
	})
}

func (s *Service) Trainers(ctx context.Context) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `
		select trainer_user_id, count(*) as session_count
		from sessions_university
		group by trainer_user_id
		order by session_count desc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []map[string]any
	for rows.Next() {
		var trainerUserID uuid.UUID
		var sessionCount int64
		if err := rows.Scan(&trainerUserID, &sessionCount); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{
			"trainer_user_id": trainerUserID,
			"session_count":   sessionCount,
		})
	}
	return items, rows.Err()
}

func (s *Service) RisksHR(ctx context.Context) (map[string]any, error) {
	counts, err := s.simpleCounts(ctx, map[string]string{
		"overdue_enrollments": `select count(*) from enrollments
			where status not in ('completed','canceled')
			  and deadline_at is not null
			  and deadline_at < now()`,
		"deadline_soon": `select count(*) from enrollments
			where status not in ('completed','canceled')
			  and deadline_at is not null
			  and deadline_at between now() and now() + interval '7 days'`,
		"inactive_learners": `select count(distinct user_id) from enrollments
			where status = 'in_progress'
			  and (last_activity_at is null or last_activity_at < now() - interval '30 days')`,
		"low_completion": `select count(*) from enrollments
			where status = 'in_progress'
			  and deadline_at is not null
			  and deadline_at < now() + interval '14 days'
			  and cast(completion_percent as decimal) < 50`,
	})
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		select e.id, e.user_id, coalesce(u.display_name, u.email), e.course_id,
		       coalesce(c.title, ''), e.deadline_at, e.completion_percent, e.last_activity_at
		from enrollments e
		join users u on u.id = e.user_id
		left join courses c on c.id = e.course_id
		where e.status not in ('completed','canceled')
		  and e.deadline_at is not null
		  and e.deadline_at < now()
		order by e.deadline_at asc
		limit 20
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var overdueItems []map[string]any
	for rows.Next() {
		var id, userID, courseID uuid.UUID
		var fullName, courseTitle, completionPercent string
		var deadlineAt time.Time
		var lastActivity *time.Time
		if err := rows.Scan(&id, &userID, &fullName, &courseID, &courseTitle, &deadlineAt, &completionPercent, &lastActivity); err != nil {
			return nil, err
		}
		overdueItems = append(overdueItems, map[string]any{
			"enrollment_id":     id,
			"user_id":           userID,
			"full_name":         fullName,
			"course_id":         courseID,
			"course_title":      courseTitle,
			"deadline_at":       deadlineAt,
			"completion_percent": completionPercent,
			"last_activity_at":  lastActivity,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	counts["overdue_items"] = overdueItems
	return counts, nil
}

func (s *Service) QueueExport(ctx context.Context, principal platformauth.Principal, format string) (map[string]any, error) {
	now := time.Now().UTC()
	jobKey := "report-export:" + format + ":" + principal.UserID.String() + ":" + now.Format("200601021504")
	if err := s.queue.Enqueue(ctx, s.db, "reports", "export_"+format, map[string]any{
		"user_id": principal.UserID,
		"format":  format,
	}, &jobKey, now); err != nil {
		return nil, err
	}
	return map[string]any{
		"queued": true,
		"format": format,
		"key":    jobKey,
	}, nil
}

func (s *Service) simpleCounts(ctx context.Context, queries map[string]string) (map[string]any, error) {
	result := make(map[string]any, len(queries))
	for key, query := range queries {
		var count int64
		if err := s.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
			return nil, err
		}
		result[key] = count
	}
	return result, nil
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func analyticsPrincipal(r *http.Request) (platformauth.Principal, error) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		return platformauth.Principal{}, httpx.Unauthorized("unauthorized", "authorization required")
	}
	return principal, nil
}

func (h *Handler) DashboardHR(w http.ResponseWriter, r *http.Request) {
	payload, err := h.service.DashboardHR(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, payload)
}

func (h *Handler) DashboardManager(w http.ResponseWriter, r *http.Request) {
	principal, err := analyticsPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	payload, err := h.service.DashboardManager(r.Context(), principal)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, payload)
}

func (h *Handler) Compliance(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.Compliance(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) ExternalRequests(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ExternalRequests(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) Budget(w http.ResponseWriter, r *http.Request) {
	payload, err := h.service.Budget(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, payload)
}

func (h *Handler) Trainers(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.Trainers(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) RisksHR(w http.ResponseWriter, r *http.Request) {
	payload, err := h.service.RisksHR(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, payload)
}

func (h *Handler) ExportExcel(w http.ResponseWriter, r *http.Request) {
	principal, err := analyticsPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	payload, err := h.service.QueueExport(r.Context(), principal, "excel")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusAccepted, payload)
}

func (h *Handler) ExportPDF(w http.ResponseWriter, r *http.Request) {
	principal, err := analyticsPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	payload, err := h.service.QueueExport(r.Context(), principal, "pdf")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusAccepted, payload)
}
