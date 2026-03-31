package external_training

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"moneyapp/backend/internal/core/common"
	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/db"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/google/uuid"
)

const defaultExternalListLimit = 20

type RequestListFilters struct {
	Scope      string
	Statuses   []string
	AssigneeID *uuid.UUID
	Pagination common.Pagination
}

type PendingApprovalItem struct {
	Request     ExternalRequest        `json:"request"`
	CurrentStep PendingApprovalStepDTO `json:"current_step"`
}

type PendingApprovalStepDTO struct {
	ID             uuid.UUID  `json:"step_id"`
	RoleCode       string     `json:"role_code"`
	DueAt          *time.Time `json:"due_at,omitempty"`
	ApproverUserID uuid.UUID  `json:"approver_user_id"`
}

func parseRequestListFilters(r *http.Request) (RequestListFilters, error) {
	scope := strings.TrimSpace(r.URL.Query().Get("scope"))
	if scope == "" {
		scope = "my"
	}
	switch scope {
	case "my", "team", "all":
	default:
		return RequestListFilters{}, httpx.BadRequest("invalid_scope", "scope must be one of my, team, all")
	}

	var assigneeID *uuid.UUID
	if raw := strings.TrimSpace(r.URL.Query().Get("assignee")); raw != "" {
		value, err := uuid.Parse(raw)
		if err != nil {
			return RequestListFilters{}, httpx.BadRequest("invalid_assignee", "assignee must be a valid UUID")
		}
		assigneeID = &value
	}

	return RequestListFilters{
		Scope:      scope,
		Statuses:   parseRepeatedStrings(r, "status"),
		AssigneeID: assigneeID,
		Pagination: common.PaginationFromRequest(r, defaultExternalListLimit),
	}, nil
}

func parseRepeatedStrings(r *http.Request, key string) []string {
	values := r.URL.Query()[key]
	if len(values) == 0 {
		return nil
	}

	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, raw := range values {
		for _, part := range strings.Split(raw, ",") {
			value := strings.TrimSpace(part)
			if value == "" {
				continue
			}
			if _, ok := seen[value]; ok {
				continue
			}
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

const externalRequestReadColumns = `
	select
		r.id,
		r.request_no,
		r.employee_user_id,
		r.department_id,
		r.title,
		r.provider_id,
		r.provider_name,
		r.course_url,
		r.program_description,
		r.planned_start_date::timestamptz,
		r.planned_end_date::timestamptz,
		r.duration_hours::text,
		r.cost_amount::text,
		r.currency,
		r.business_goal,
		r.employee_comment,
		r.manager_comment,
		r.hr_comment,
		r.status,
		r.calendar_conflict_status,
		r.budget_check_status,
		r.current_approval_step_id,
		r.approved_at,
		r.rejected_at,
		r.sent_to_revision_at,
		r.training_started_at,
		r.training_completed_at,
		r.certificate_uploaded_at,
		r.created_at,
		r.updated_at,
		coalesce(nullif(trim(concat_ws(' ', ep.last_name, ep.first_name, ep.middle_name)), ''), '') as employee_full_name,
		u.email as employee_email,
		coalesce(d.name, '') as department_name,
		coalesce(s.status, '') as current_approval_status,
		coalesce(s.role_code, '') as current_approval_role_code,
		s.due_at,
		s.approver_user_id::text,
		coalesce(nullif(trim(concat_ws(' ', ap.last_name, ap.first_name, ap.middle_name)), ''), '') as current_approver_full_name
`

const externalRequestReadJoins = `
	from external_course_requests r
	join users u on u.id = r.employee_user_id
	left join employee_profiles ep on ep.user_id = r.employee_user_id
	left join departments d on d.id = r.department_id
	left join approval_steps s on s.id = r.current_approval_step_id
	left join employee_profiles ap on ap.user_id = s.approver_user_id
`

func canListExternalScope(principal platformauth.Principal, scope string) bool {
	switch scope {
	case "my":
		return canReadOwnExternalRequests(principal)
	case "team":
		return canViewTeamExternalRequests(principal)
	case "all":
		return canViewAllExternalRequests(principal)
	default:
		return false
	}
}

type externalRequestRowScanner interface {
	Scan(dest ...any) error
}

func scanExternalRequest(scanner externalRequestRowScanner) (ExternalRequest, error) {
	var item ExternalRequest
	var (
		currentApprovalDueAt sql.NullTime
		currentApproverID    sql.NullString
	)
	err := scanner.Scan(
		&item.ID,
		&item.RequestNo,
		&item.EmployeeUserID,
		&item.DepartmentID,
		&item.Title,
		&item.ProviderID,
		&item.ProviderName,
		&item.CourseURL,
		&item.ProgramDescription,
		&item.PlannedStartDate,
		&item.PlannedEndDate,
		&item.DurationHours,
		&item.CostAmount,
		&item.Currency,
		&item.BusinessGoal,
		&item.EmployeeComment,
		&item.ManagerComment,
		&item.HRComment,
		&item.Status,
		&item.CalendarConflictStatus,
		&item.BudgetCheckStatus,
		&item.CurrentApprovalStepID,
		&item.ApprovedAt,
		&item.RejectedAt,
		&item.SentToRevisionAt,
		&item.TrainingStartedAt,
		&item.TrainingCompletedAt,
		&item.CertificateUploadedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.EmployeeFullName,
		&item.EmployeeEmail,
		&item.DepartmentName,
		&item.CurrentApprovalStatus,
		&item.CurrentApprovalRole,
		&currentApprovalDueAt,
		&currentApproverID,
		&item.CurrentApproverName,
	)
	if err != nil {
		return ExternalRequest{}, err
	}
	if currentApprovalDueAt.Valid {
		value := currentApprovalDueAt.Time
		item.CurrentApprovalDueAt = &value
	}
	if currentApproverID.Valid {
		value, err := uuid.Parse(currentApproverID.String)
		if err != nil {
			return ExternalRequest{}, err
		}
		item.CurrentApproverUserID = &value
	}
	return item, nil
}

func (r *Repository) ListRequests(ctx context.Context, principal platformauth.Principal, filters RequestListFilters, exec ...db.DBTX) ([]ExternalRequest, error) {
	args := make([]any, 0, 8)
	where := make([]string, 0, 6)

	query := strings.Builder{}
	query.WriteString(externalRequestReadColumns)
	query.WriteString(externalRequestReadJoins)
	query.WriteString(`
		left join manager_relations mr on mr.employee_user_id = r.employee_user_id and mr.is_primary = true
	`)

	switch filters.Scope {
	case "my":
		args = append(args, principal.UserID)
		where = append(where, fmt.Sprintf("r.employee_user_id = $%d", len(args)))
	case "team":
		args = append(args, principal.UserID)
		where = append(where, fmt.Sprintf("mr.manager_user_id = $%d", len(args)))
	}

	if len(filters.Statuses) > 0 {
		placeholders := make([]string, 0, len(filters.Statuses))
		for _, value := range filters.Statuses {
			args = append(args, value)
			placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)))
		}
		where = append(where, fmt.Sprintf("r.status in (%s)", strings.Join(placeholders, ", ")))
	}

	if filters.AssigneeID != nil {
		args = append(args, *filters.AssigneeID)
		where = append(where, fmt.Sprintf("s.approver_user_id = $%d", len(args)))
		where = append(where, "s.status = 'pending'")
	}

	if len(where) > 0 {
		query.WriteString(" where ")
		query.WriteString(strings.Join(where, " and "))
	}
	args = append(args, filters.Pagination.Limit, filters.Pagination.Offset)
	query.WriteString(fmt.Sprintf(" order by r.created_at desc limit $%d offset $%d", len(args)-1, len(args)))

	rows, err := r.base(exec...).QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ExternalRequest
	for rows.Next() {
		item, err := scanExternalRequest(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ListPendingApprovals(ctx context.Context, approverID uuid.UUID, pagination common.Pagination, exec ...db.DBTX) ([]PendingApprovalItem, error) {
	rows, err := r.base(exec...).QueryContext(ctx, externalRequestReadColumns+externalRequestReadJoins+`
		where s.approver_user_id = $1 and s.status = 'pending'
		order by s.due_at asc nulls last, r.created_at desc
		limit $2 offset $3
	`, approverID, pagination.Limit, pagination.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []PendingApprovalItem
	for rows.Next() {
		request, err := scanExternalRequest(rows)
		if err != nil {
			return nil, err
		}
		if request.CurrentApprovalStepID == nil || request.CurrentApproverUserID == nil {
			return nil, fmt.Errorf("pending approval request %s is missing current step", request.ID)
		}
		item := PendingApprovalItem{
			Request: request,
			CurrentStep: PendingApprovalStepDTO{
				ID:             *request.CurrentApprovalStepID,
				RoleCode:       request.CurrentApprovalRole,
				DueAt:          request.CurrentApprovalDueAt,
				ApproverUserID: *request.CurrentApproverUserID,
			},
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) List(ctx context.Context, principal platformauth.Principal, filters RequestListFilters) ([]ExternalRequest, error) {
	if !canListExternalScope(principal, filters.Scope) {
		return nil, httpx.Forbidden("forbidden", "permission denied")
	}
	return s.repo.ListRequests(ctx, principal, filters)
}

func (s *Service) ListPendingApprovals(ctx context.Context, principal platformauth.Principal, pagination common.Pagination) ([]PendingApprovalItem, error) {
	if !canViewTeamExternalRequests(principal) && !canViewAllExternalRequests(principal) {
		return nil, httpx.Forbidden("forbidden", "permission denied")
	}
	return s.repo.ListPendingApprovals(ctx, principal.UserID, pagination)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	principal, err := externalPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	filters, err := parseRequestListFilters(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	items, err := h.service.List(r.Context(), principal, filters)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) PendingApprovals(w http.ResponseWriter, r *http.Request) {
	principal, err := externalPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	items, err := h.service.ListPendingApprovals(r.Context(), principal, common.PaginationFromRequest(r, defaultExternalListLimit))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}
