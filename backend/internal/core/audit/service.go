package audit

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

func (s *Service) Record(ctx context.Context, userID uuid.UUID, action, entityType string, entityID *uuid.UUID, meta map[string]any) error {
	payload, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	return s.repo.Create(ctx, Event{
		ID:         uuid.New(),
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Meta:       payload,
		CreatedAt:  s.clock.Now(),
	})
}
