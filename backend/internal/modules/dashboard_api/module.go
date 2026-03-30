package dashboard_api

import (
	"context"
	"database/sql"
	"net/http"

	"moneyapp/backend/internal/core/common"
	calendarread "moneyapp/backend/internal/modules/calendar"
	courserequests "moneyapp/backend/internal/modules/course_requests"
	externaltraining "moneyapp/backend/internal/modules/external_training"
	learningplan "moneyapp/backend/internal/modules/learning_plan"
	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/google/uuid"
)

type EmployeeStats struct {
	ActiveEnrollments    int `json:"active_enrollments"`
	RecommendedCourses   int `json:"recommended_courses"`
	OpenExternalRequests int `json:"open_external_requests"`
}

type TeamPreviewItem struct {
	UserID        uuid.UUID  `json:"user_id"`
	FirstName     string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	PositionTitle *string    `json:"position_title,omitempty"`
	DepartmentID  *uuid.UUID `json:"department_id,omitempty"`
}

type ManagerStats struct {
	TeamSize                 int `json:"team_size"`
	PendingExternalApprovals int `json:"pending_external_approvals"`
	TeamExternalRequests     int `json:"team_external_requests"`
	TeamCourseRequests       int `json:"team_course_requests"`
}

type EmployeeDashboard struct {
	Stats                   EmployeeStats                        `json:"stats"`
	UpcomingEvents          []calendarread.UpcomingCalendarEvent `json:"upcoming_events"`
	RecommendedCourses      []learningplan.RecommendedCourseItem `json:"recommended_courses"`
	LearningPlan            learningplan.MyLearningPlan          `json:"learning_plan"`
	ExternalRequestsPreview []externaltraining.ExternalRequest   `json:"external_requests_preview"`
}

type ManagerDashboard struct {
	Stats                    ManagerStats                           `json:"stats"`
	TeamPreview              []TeamPreviewItem                      `json:"team_preview"`
	PendingExternalApprovals []externaltraining.PendingApprovalItem `json:"pending_external_approvals"`
	TeamExternalRequests     []externaltraining.ExternalRequest     `json:"team_external_requests"`
	TeamCourseRequests       []courserequests.CourseRequest         `json:"team_course_requests"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(database *sql.DB) *Repository {
	return &Repository{db: database}
}

func (r *Repository) countActiveEnrollments(ctx context.Context, userID uuid.UUID) (int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, `
		select count(*)
		from enrollments
		where user_id = $1 and status in ('not_started', 'in_progress')
	`, userID).Scan(&total)
	return total, err
}

func (r *Repository) countOpenExternalRequests(ctx context.Context, userID uuid.UUID) (int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, `
		select count(*)
		from external_course_requests
		where employee_user_id = $1
		  and status not in ('rejected', 'closed', 'canceled', 'completed')
	`, userID).Scan(&total)
	return total, err
}

func (r *Repository) teamPreview(ctx context.Context, managerID uuid.UUID, limit int) ([]TeamPreviewItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		select ep.user_id, ep.first_name, ep.last_name, ep.position_title, ep.department_id
		from manager_relations mr
		join employee_profiles ep on ep.user_id = mr.employee_user_id
		where mr.manager_user_id = $1 and mr.is_primary = true
		order by ep.last_name asc, ep.first_name asc
		limit $2
	`, managerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []TeamPreviewItem
	for rows.Next() {
		var item TeamPreviewItem
		if err := rows.Scan(&item.UserID, &item.FirstName, &item.LastName, &item.PositionTitle, &item.DepartmentID); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) countTeamSize(ctx context.Context, managerID uuid.UUID) (int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, `
		select count(*)
		from manager_relations
		where manager_user_id = $1 and is_primary = true
	`, managerID).Scan(&total)
	return total, err
}

func (r *Repository) countPendingExternalApprovals(ctx context.Context, approverID uuid.UUID) (int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, `
		select count(*)
		from approval_steps s
		join external_course_requests r on r.current_approval_step_id = s.id
		where s.approver_user_id = $1 and s.status = 'pending'
	`, approverID).Scan(&total)
	return total, err
}

func (r *Repository) countTeamExternalRequests(ctx context.Context, managerID uuid.UUID) (int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, `
		select count(*)
		from external_course_requests r
		join manager_relations mr on mr.employee_user_id = r.employee_user_id and mr.is_primary = true
		where mr.manager_user_id = $1
	`, managerID).Scan(&total)
	return total, err
}

type Service struct {
	repo                 *Repository
	calendarService      *calendarread.Service
	learningPlanService  *learningplan.Service
	externalService      *externaltraining.Service
	courseRequestService *courserequests.Service
}

func NewService(repo *Repository, calendarService *calendarread.Service, learningPlanService *learningplan.Service, externalService *externaltraining.Service, courseRequestService *courserequests.Service) *Service {
	return &Service{
		repo:                 repo,
		calendarService:      calendarService,
		learningPlanService:  learningPlanService,
		externalService:      externalService,
		courseRequestService: courseRequestService,
	}
}

func (s *Service) Employee(ctx context.Context, principal platformauth.Principal) (EmployeeDashboard, error) {
	activeEnrollments, err := s.repo.countActiveEnrollments(ctx, principal.UserID)
	if err != nil {
		return EmployeeDashboard{}, err
	}
	openExternalRequests, err := s.repo.countOpenExternalRequests(ctx, principal.UserID)
	if err != nil {
		return EmployeeDashboard{}, err
	}
	upcomingEvents, err := s.calendarService.ListUpcoming(ctx, principal, 5)
	if err != nil {
		return EmployeeDashboard{}, err
	}
	learningPlan, err := s.learningPlanService.MyPlan(ctx, principal)
	if err != nil {
		return EmployeeDashboard{}, err
	}
	recommendedCourses, err := s.learningPlanService.RecommendedCourses(ctx, principal, common.Pagination{Limit: 6, Offset: 0})
	if err != nil {
		return EmployeeDashboard{}, err
	}
	externalRequests, err := s.externalService.List(ctx, principal, externaltraining.RequestListFilters{
		Scope: "my",
		Pagination: common.Pagination{
			Limit:  5,
			Offset: 0,
		},
	})
	if err != nil {
		return EmployeeDashboard{}, err
	}
	learningPlan.Recommended = recommendedCourses
	return EmployeeDashboard{
		Stats: EmployeeStats{
			ActiveEnrollments:    activeEnrollments,
			RecommendedCourses:   learningPlan.Summary.Recommended,
			OpenExternalRequests: openExternalRequests,
		},
		UpcomingEvents:          upcomingEvents,
		RecommendedCourses:      recommendedCourses,
		LearningPlan:            learningPlan,
		ExternalRequestsPreview: externalRequests,
	}, nil
}

func canViewManagerDashboard(principal platformauth.Principal) bool {
	return principal.HasPermission("external_requests.approve_manager") ||
		principal.HasPermission("external_requests.approve_hr") ||
		principal.HasPermission("external_requests.read_all")
}

func (s *Service) Manager(ctx context.Context, principal platformauth.Principal) (ManagerDashboard, error) {
	if !canViewManagerDashboard(principal) {
		return ManagerDashboard{}, httpx.Forbidden("forbidden", "permission denied")
	}

	teamSize, err := s.repo.countTeamSize(ctx, principal.UserID)
	if err != nil {
		return ManagerDashboard{}, err
	}
	pendingApprovalsCount, err := s.repo.countPendingExternalApprovals(ctx, principal.UserID)
	if err != nil {
		return ManagerDashboard{}, err
	}
	teamExternalRequestsCount, err := s.repo.countTeamExternalRequests(ctx, principal.UserID)
	if err != nil {
		return ManagerDashboard{}, err
	}
	teamPreview, err := s.repo.teamPreview(ctx, principal.UserID, 5)
	if err != nil {
		return ManagerDashboard{}, err
	}
	pendingApprovals, err := s.externalService.ListPendingApprovals(ctx, principal, common.Pagination{Limit: 5, Offset: 0})
	if err != nil {
		return ManagerDashboard{}, err
	}
	teamExternalRequests, err := s.externalService.List(ctx, principal, externaltraining.RequestListFilters{
		Scope: "team",
		Pagination: common.Pagination{
			Limit:  5,
			Offset: 0,
		},
	})
	if err != nil {
		return ManagerDashboard{}, err
	}
	teamCourseRequests, err := s.courseRequestService.List(ctx, principal)
	if err != nil {
		return ManagerDashboard{}, err
	}
	teamCourseRequestsTotal := len(teamCourseRequests)
	if len(teamCourseRequests) > 5 {
		teamCourseRequests = teamCourseRequests[:5]
	}

	return ManagerDashboard{
		Stats: ManagerStats{
			TeamSize:                 teamSize,
			PendingExternalApprovals: pendingApprovalsCount,
			TeamExternalRequests:     teamExternalRequestsCount,
			TeamCourseRequests:       teamCourseRequestsTotal,
		},
		TeamPreview:              teamPreview,
		PendingExternalApprovals: pendingApprovals,
		TeamExternalRequests:     teamExternalRequests,
		TeamCourseRequests:       teamCourseRequests,
	}, nil
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func dashboardPrincipal(r *http.Request) (platformauth.Principal, error) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		return platformauth.Principal{}, httpx.Unauthorized("unauthorized", "authorization required")
	}
	return principal, nil
}

func (h *Handler) Employee(w http.ResponseWriter, r *http.Request) {
	principal, err := dashboardPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	payload, err := h.service.Employee(r.Context(), principal)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, payload)
}

func (h *Handler) Manager(w http.ResponseWriter, r *http.Request) {
	principal, err := dashboardPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	payload, err := h.service.Manager(r.Context(), principal)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, payload)
}
