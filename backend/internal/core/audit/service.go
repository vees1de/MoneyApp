package audit

import (
	"context"
	"encoding/json"
	"log/slog"

	"moneyapp/backend/internal/platform/clock"

	"github.com/google/uuid"
)

type Publisher interface {
	PublishJSON(ctx context.Context, topic, key string, payload any) error
}

type Service struct {
	repo      *Repository
	clock     clock.Clock
	logger    *slog.Logger
	publisher Publisher
	topic     string
}

func NewService(repo *Repository, clock clock.Clock, logger *slog.Logger, publisher Publisher, topic string) *Service {
	return &Service{
		repo:      repo,
		clock:     clock,
		logger:    logger,
		publisher: publisher,
		topic:     topic,
	}
}

func (s *Service) Record(ctx context.Context, userID uuid.UUID, action, entityType string, entityID *uuid.UUID, meta map[string]any) error {
	payload, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	event := Event{
		ID:         uuid.New(),
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Meta:       payload,
		CreatedAt:  s.clock.Now(),
	}
	if err := s.repo.Create(ctx, event); err != nil {
		return err
	}

	if s.publisher != nil && s.topic != "" {
		message := map[string]any{
			"id":          event.ID,
			"user_id":     event.UserID,
			"action":      event.Action,
			"entity_type": event.EntityType,
			"entity_id":   event.EntityID,
			"meta":        meta,
			"created_at":  event.CreatedAt,
		}
		if err := s.publisher.PublishJSON(ctx, s.topic, event.UserID.String(), message); err != nil && s.logger != nil {
			s.logger.Error("publish audit event", "error", err, "action", action, "entity_type", entityType)
		}
	}

	return nil
}
