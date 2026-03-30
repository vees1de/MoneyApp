package learning_plan

import (
	"context"
	"database/sql"
	"net/http"
	"sort"
	"time"

	"moneyapp/backend/internal/core/common"
	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/google/uuid"
)

const defaultRecommendationLimit = 10

type LearningPlanItem struct {
	EnrollmentID        uuid.UUID  `json:"enrollment_id"`
	CourseID            uuid.UUID  `json:"course_id"`
	AssignmentID        *uuid.UUID `json:"assignment_id,omitempty"`
	Source              string     `json:"source"`
	Status              string     `json:"status"`
	Title               string     `json:"title"`
	ShortDescription    *string    `json:"short_description,omitempty"`
	DeadlineAt          *time.Time `json:"deadline_at,omitempty"`
	StartedAt           *time.Time `json:"started_at,omitempty"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	CompletionPercent   string     `json:"completion_percent"`
	IsMandatory         bool       `json:"is_mandatory"`
	Reason              *string    `json:"reason,omitempty"`
	EnrollmentCreatedAt time.Time  `json:"enrollment_created_at"`
}

type RecommendedCourseItem struct {
	CourseID            uuid.UUID  `json:"course_id"`
	EnrollmentID        uuid.UUID  `json:"enrollment_id"`
	AssignmentID        *uuid.UUID `json:"assignment_id,omitempty"`
	Title               string     `json:"title"`
	ShortDescription    *string    `json:"short_description,omitempty"`
	Status              string     `json:"status"`
	DeadlineAt          *time.Time `json:"deadline_at,omitempty"`
	CompletionPercent   string     `json:"completion_percent"`
	Reason              *string    `json:"reason,omitempty"`
	EnrollmentCreatedAt time.Time  `json:"enrollment_created_at"`
}

type LearningPlanSummary struct {
	Total             int `json:"total"`
	InProgress        int `json:"in_progress"`
	Upcoming          int `json:"upcoming"`
	CompletedRecently int `json:"completed_recently"`
	Recommended       int `json:"recommended"`
}

type MyLearningPlan struct {
	Summary           LearningPlanSummary     `json:"summary"`
	InProgress        []LearningPlanItem      `json:"in_progress"`
	Upcoming          []LearningPlanItem      `json:"upcoming"`
	CompletedRecently []LearningPlanItem      `json:"completed_recently"`
	Recommended       []RecommendedCourseItem `json:"recommended"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(database *sql.DB) *Repository {
	return &Repository{db: database}
}

func (r *Repository) listUserLearningItems(ctx context.Context, userID uuid.UUID) ([]LearningPlanItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		select e.id, e.course_id, e.assignment_id, e.source, e.status, c.title, c.short_description,
		       e.deadline_at, e.started_at, e.completed_at, e.completion_percent::text, e.is_mandatory,
		       a.reason, e.created_at
		from enrollments e
		join courses c on c.id = e.course_id
		left join course_assignments a on a.id = e.assignment_id
		where e.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []LearningPlanItem
	for rows.Next() {
		var item LearningPlanItem
		if err := rows.Scan(&item.EnrollmentID, &item.CourseID, &item.AssignmentID, &item.Source, &item.Status, &item.Title, &item.ShortDescription,
			&item.DeadlineAt, &item.StartedAt, &item.CompletedAt, &item.CompletionPercent, &item.IsMandatory, &item.Reason, &item.EnrollmentCreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) MyPlan(ctx context.Context, principal platformauth.Principal) (MyLearningPlan, error) {
	items, err := s.repo.listUserLearningItems(ctx, principal.UserID)
	if err != nil {
		return MyLearningPlan{}, err
	}

	plan := MyLearningPlan{}
	for _, item := range items {
		plan.Summary.Total++
		switch item.Status {
		case "in_progress":
			plan.Summary.InProgress++
			plan.InProgress = append(plan.InProgress, item)
		case "not_started":
			plan.Summary.Upcoming++
			plan.Upcoming = append(plan.Upcoming, item)
		case "completed":
			plan.Summary.CompletedRecently++
			plan.CompletedRecently = append(plan.CompletedRecently, item)
		}
		if item.Source == "assignment" && !item.IsMandatory && (item.Status == "not_started" || item.Status == "in_progress") {
			plan.Summary.Recommended++
			plan.Recommended = append(plan.Recommended, RecommendedCourseItem{
				CourseID:            item.CourseID,
				EnrollmentID:        item.EnrollmentID,
				AssignmentID:        item.AssignmentID,
				Title:               item.Title,
				ShortDescription:    item.ShortDescription,
				Status:              item.Status,
				DeadlineAt:          item.DeadlineAt,
				CompletionPercent:   item.CompletionPercent,
				Reason:              item.Reason,
				EnrollmentCreatedAt: item.EnrollmentCreatedAt,
			})
		}
	}

	sort.Slice(plan.InProgress, func(i, j int) bool {
		return comparePlanItems(plan.InProgress[i], plan.InProgress[j])
	})
	sort.Slice(plan.Upcoming, func(i, j int) bool {
		return comparePlanItems(plan.Upcoming[i], plan.Upcoming[j])
	})
	sort.Slice(plan.CompletedRecently, func(i, j int) bool {
		left := plan.CompletedRecently[i].CompletedAt
		right := plan.CompletedRecently[j].CompletedAt
		if left == nil || right == nil {
			return plan.CompletedRecently[i].EnrollmentCreatedAt.After(plan.CompletedRecently[j].EnrollmentCreatedAt)
		}
		return left.After(*right)
	})
	sort.Slice(plan.Recommended, func(i, j int) bool {
		return compareRecommendedItems(plan.Recommended[i], plan.Recommended[j])
	})
	plan.Recommended = dedupeRecommendations(plan.Recommended)

	return plan, nil
}

func comparePlanItems(left, right LearningPlanItem) bool {
	if left.DeadlineAt == nil {
		return false
	}
	if right.DeadlineAt == nil {
		return true
	}
	if left.DeadlineAt.Equal(*right.DeadlineAt) {
		return left.EnrollmentCreatedAt.After(right.EnrollmentCreatedAt)
	}
	return left.DeadlineAt.Before(*right.DeadlineAt)
}

func compareRecommendedItems(left, right RecommendedCourseItem) bool {
	if left.DeadlineAt == nil {
		return false
	}
	if right.DeadlineAt == nil {
		return true
	}
	if left.DeadlineAt.Equal(*right.DeadlineAt) {
		return left.EnrollmentCreatedAt.After(right.EnrollmentCreatedAt)
	}
	return left.DeadlineAt.Before(*right.DeadlineAt)
}

func dedupeRecommendations(items []RecommendedCourseItem) []RecommendedCourseItem {
	result := make([]RecommendedCourseItem, 0, len(items))
	seen := make(map[uuid.UUID]struct{}, len(items))
	for _, item := range items {
		if _, ok := seen[item.CourseID]; ok {
			continue
		}
		seen[item.CourseID] = struct{}{}
		result = append(result, item)
	}
	return result
}

func (s *Service) RecommendedCourses(ctx context.Context, principal platformauth.Principal, pagination common.Pagination) ([]RecommendedCourseItem, error) {
	plan, err := s.MyPlan(ctx, principal)
	if err != nil {
		return nil, err
	}
	items := plan.Recommended
	if pagination.Offset >= len(items) {
		return []RecommendedCourseItem{}, nil
	}
	end := pagination.Offset + pagination.Limit
	if end > len(items) {
		end = len(items)
	}
	return items[pagination.Offset:end], nil
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func planPrincipal(r *http.Request) (platformauth.Principal, error) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		return platformauth.Principal{}, httpx.Unauthorized("unauthorized", "authorization required")
	}
	return principal, nil
}

func (h *Handler) MyPlan(w http.ResponseWriter, r *http.Request) {
	principal, err := planPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	payload, err := h.service.MyPlan(r.Context(), principal)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, payload)
}

func (h *Handler) RecommendedCourses(w http.ResponseWriter, r *http.Request) {
	principal, err := planPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	items, err := h.service.RecommendedCourses(r.Context(), principal, common.PaginationFromRequest(r, defaultRecommendationLimit))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}
