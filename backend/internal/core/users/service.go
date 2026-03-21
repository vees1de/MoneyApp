package users

import (
	"context"
	"database/sql"

	"moneyapp/backend/internal/platform/httpx"

	"github.com/google/uuid"
)

type Service struct {
	db   *sql.DB
	repo *Repository
}

func NewService(database *sql.DB, repo *Repository) *Service {
	return &Service{
		db:   database,
		repo: repo,
	}
}

func (s *Service) GetByID(ctx context.Context, userID uuid.UUID) (User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if IsNotFound(err) {
		return User{}, httpx.NotFound("user_not_found", "user not found")
	}
	return user, err
}

func (s *Service) GetProfile(ctx context.Context, userID uuid.UUID) (User, error) {
	return s.GetByID(ctx, userID)
}

func (s *Service) UpdatePreferences(ctx context.Context, userID uuid.UUID, request UpdatePreferencesRequest) (User, error) {
	user, err := s.repo.UpdatePreferences(ctx, userID, request)
	if IsNotFound(err) {
		return User{}, httpx.NotFound("user_not_found", "user not found")
	}
	return user, err
}
