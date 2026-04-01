package employees_stats

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/httpx"
)

// --- Models ---

type EmployeeLearningStats struct {
	UserID          uuid.UUID  `json:"user_id"`
	FirstName       string     `json:"first_name"`
	LastName        string     `json:"last_name"`
	MiddleName      *string    `json:"middle_name"`
	FullName        string     `json:"full_name"`
	Email           string     `json:"email"`
	PositionTitle   *string    `json:"position_title"`
	DepartmentID    *uuid.UUID `json:"department_id"`
	DepartmentName  *string    `json:"department_name"`
	InProgressCount int        `json:"in_progress_count"`
	CompletedCount  int        `json:"completed_count"`
	OverdueCount    int        `json:"overdue_count"`
}

type LearningStatsResponse struct {
	Items  []EmployeeLearningStats `json:"items"`
	Total  int                     `json:"total"`
	Limit  int                     `json:"limit"`
	Offset int                     `json:"offset"`
}

type LearningStatsFilter struct {
	Scope        string
	DepartmentID *uuid.UUID
	Search       string
	Limit        int
	Offset       int
	ManagerID    uuid.UUID
}

// --- Repository ---

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) LearningStats(ctx context.Context, f LearningStatsFilter) ([]EmployeeLearningStats, int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	conditions = append(conditions, "ep.employment_status = 'active'")

	if f.Scope == "team" {
		conditions = append(conditions, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM manager_relations mr WHERE mr.employee_user_id = ep.user_id AND mr.manager_user_id = $%d)", argIdx))
		args = append(args, f.ManagerID)
		argIdx++
	}

	if f.DepartmentID != nil {
		conditions = append(conditions, fmt.Sprintf("ep.department_id = $%d", argIdx))
		args = append(args, *f.DepartmentID)
		argIdx++
	}

	if f.Search != "" {
		searchPattern := "%" + strings.ToLower(f.Search) + "%"
		conditions = append(conditions, fmt.Sprintf(
			"(LOWER(ep.first_name || ' ' || ep.last_name || ' ' || COALESCE(ep.middle_name, '')) LIKE $%d OR LOWER(u.email) LIKE $%d)",
			argIdx, argIdx))
		args = append(args, searchPattern)
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM employee_profiles ep
		JOIN users u ON u.id = ep.user_id
		%s`, where)

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []EmployeeLearningStats{}, 0, nil
	}

	dataQuery := fmt.Sprintf(`
		SELECT
			ep.user_id,
			ep.first_name,
			ep.last_name,
			ep.middle_name,
			u.email,
			ep.position_title,
			ep.department_id,
			d.name AS department_name,
			COALESCE(SUM(CASE WHEN e.status IN ('not_started', 'in_progress') THEN 1 ELSE 0 END), 0) AS in_progress_count,
			COALESCE(SUM(CASE WHEN e.status = 'completed' THEN 1 ELSE 0 END), 0) AS completed_count,
			COALESCE(SUM(CASE WHEN e.deadline_at IS NOT NULL AND e.deadline_at < NOW() AND e.status NOT IN ('completed', 'canceled') THEN 1 ELSE 0 END), 0) AS overdue_count
		FROM employee_profiles ep
		JOIN users u ON u.id = ep.user_id
		LEFT JOIN departments d ON d.id = ep.department_id
		LEFT JOIN enrollments e ON e.user_id = ep.user_id
		%s
		GROUP BY ep.user_id, ep.first_name, ep.last_name, ep.middle_name, u.email, ep.position_title, ep.department_id, d.name
		ORDER BY ep.last_name, ep.first_name
		LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1)

	args = append(args, f.Limit, f.Offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []EmployeeLearningStats
	for rows.Next() {
		var item EmployeeLearningStats
		if err := rows.Scan(
			&item.UserID,
			&item.FirstName,
			&item.LastName,
			&item.MiddleName,
			&item.Email,
			&item.PositionTitle,
			&item.DepartmentID,
			&item.DepartmentName,
			&item.InProgressCount,
			&item.CompletedCount,
			&item.OverdueCount,
		); err != nil {
			return nil, 0, err
		}

		lastName := item.LastName
		firstName := item.FirstName
		middleName := ""
		if item.MiddleName != nil {
			middleName = " " + *item.MiddleName
		}
		item.FullName = lastName + " " + firstName + middleName

		items = append(items, item)
	}

	if items == nil {
		items = []EmployeeLearningStats{}
	}

	return items, total, rows.Err()
}

// --- Service ---

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) LearningStats(ctx context.Context, f LearningStatsFilter) (*LearningStatsResponse, error) {
	items, total, err := s.repo.LearningStats(ctx, f)
	if err != nil {
		return nil, err
	}
	return &LearningStatsResponse{
		Items:  items,
		Total:  total,
		Limit:  f.Limit,
		Offset: f.Offset,
	}, nil
}

// --- Handler ---

type Handler struct {
	service   *Service
	validator *validator.Validate
}

func NewHandler(service *Service, validator *validator.Validate) *Handler {
	return &Handler{service: service, validator: validator}
}

func (h *Handler) LearningStats(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}

	scope := r.URL.Query().Get("scope")
	if scope == "" {
		scope = "team"
	}
	if scope != "team" && scope != "all" {
		httpx.WriteError(w, httpx.BadRequest("invalid_scope", "scope must be 'team' or 'all'"))
		return
	}

	if scope == "all" && !principal.HasPermission("analytics.read_hr") && !principal.HasPermission("users.read") {
		httpx.WriteError(w, httpx.Forbidden("forbidden", "permission denied"))
		return
	}

	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil || parsed < 1 {
			httpx.WriteError(w, httpx.BadRequest("invalid_limit", "limit must be a positive integer"))
			return
		}
		if parsed > 200 {
			parsed = 200
		}
		limit = parsed
	}

	offset := 0
	if v := r.URL.Query().Get("offset"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil || parsed < 0 {
			httpx.WriteError(w, httpx.BadRequest("invalid_offset", "offset must be a non-negative integer"))
			return
		}
		offset = parsed
	}

	var departmentID *uuid.UUID
	if v := r.URL.Query().Get("department_id"); v != "" {
		parsed, err := uuid.Parse(v)
		if err != nil {
			httpx.WriteError(w, httpx.BadRequest("invalid_department_id", "department_id must be a valid UUID"))
			return
		}
		departmentID = &parsed
	}

	search := r.URL.Query().Get("search")

	filter := LearningStatsFilter{
		Scope:        scope,
		DepartmentID: departmentID,
		Search:       search,
		Limit:        limit,
		Offset:       offset,
		ManagerID:    principal.UserID,
	}

	result, err := h.service.LearningStats(r.Context(), filter)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result)
}
