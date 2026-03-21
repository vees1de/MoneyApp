package categories

import (
	"context"
	"database/sql"
	"errors"

	"moneyapp/backend/internal/core/audit"
	"moneyapp/backend/internal/platform/clock"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/google/uuid"
)

type Service struct {
	repo  *Repository
	audit *audit.Service
	clock clock.Clock
}

func NewService(repo *Repository, auditService *audit.Service, clock clock.Clock) *Service {
	return &Service{
		repo:  repo,
		audit: auditService,
		clock: clock,
	}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]Category, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, request CreateCategoryRequest) (Category, error) {
	now := s.clock.Now()
	category := Category{
		ID:         uuid.New(),
		UserID:     &userID,
		Kind:       request.Kind,
		Name:       request.Name,
		Color:      request.Color,
		Icon:       request.Icon,
		ParentID:   request.ParentID,
		IsSystem:   false,
		IsArchived: false,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.repo.Create(ctx, category); err != nil {
		return Category{}, err
	}

	if err := s.audit.Record(ctx, userID, "categories.create", "category", &category.ID, map[string]any{
		"name": category.Name,
		"kind": category.Kind,
	}); err != nil {
		return Category{}, err
	}

	return category, nil
}

func (s *Service) Update(ctx context.Context, userID, categoryID uuid.UUID, request UpdateCategoryRequest) (Category, error) {
	category, err := s.repo.GetByID(ctx, userID, categoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Category{}, httpx.NotFound("category_not_found", "category not found")
		}
		return Category{}, err
	}
	if category.IsSystem {
		return Category{}, httpx.Forbidden("system_category_immutable", "system categories cannot be modified")
	}

	if request.Name != nil {
		category.Name = *request.Name
	}
	if request.Color != nil {
		category.Color = request.Color
	}
	if request.Icon != nil {
		category.Icon = request.Icon
	}
	if request.ParentID != nil {
		category.ParentID = request.ParentID
	}
	if request.IsArchived != nil {
		category.IsArchived = *request.IsArchived
	}
	category.UpdatedAt = s.clock.Now()

	if err := s.repo.Update(ctx, category); err != nil {
		return Category{}, err
	}
	if err := s.audit.Record(ctx, userID, "categories.update", "category", &category.ID, map[string]any{
		"is_archived": category.IsArchived,
	}); err != nil {
		return Category{}, err
	}

	return category, nil
}

func (s *Service) Archive(ctx context.Context, userID, categoryID uuid.UUID) error {
	_, err := s.Update(ctx, userID, categoryID, UpdateCategoryRequest{IsArchived: boolPtr(true)})
	return err
}

func boolPtr(v bool) *bool {
	return &v
}
