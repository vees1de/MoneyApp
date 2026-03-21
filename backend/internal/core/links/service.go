package links

import (
	"context"
	"encoding/json"

	"moneyapp/backend/internal/platform/clock"

	"github.com/google/uuid"
)

type Service struct {
	repo  *Repository
	clock clock.Clock
}

func NewService(repo *Repository, clock clock.Clock) *Service {
	return &Service{
		repo:  repo,
		clock: clock,
	}
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, request CreateLinkRequest) (EntityLink, error) {
	meta, err := json.Marshal(request.Meta)
	if err != nil {
		return EntityLink{}, err
	}

	link := EntityLink{
		ID:         uuid.New(),
		UserID:     userID,
		SourceType: request.SourceType,
		SourceID:   request.SourceID,
		TargetType: request.TargetType,
		TargetID:   request.TargetID,
		Relation:   request.Relation,
		Meta:       meta,
		CreatedAt:  s.clock.Now(),
	}
	if err := s.repo.Create(ctx, link); err != nil {
		return EntityLink{}, err
	}

	return link, nil
}

func (s *Service) ListByEntity(ctx context.Context, userID uuid.UUID, query ListByEntityQuery) ([]EntityLink, error) {
	return s.repo.ListByEntity(ctx, userID, query.EntityType, query.EntityID)
}
