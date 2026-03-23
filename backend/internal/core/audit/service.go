package audit

import (
	"context"
	"encoding/json"
	"log/slog"

	"moneyapp/backend/internal/middleware"
	platformauth "moneyapp/backend/internal/platform/auth"
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
	return s.RecordChange(ctx, RecordInput{
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Meta:       meta,
	})
}

func (s *Service) RecordChange(ctx context.Context, input RecordInput) error {
	payload, err := json.Marshal(input.Meta)
	if err != nil {
		return err
	}

	changeSet, err := json.Marshal(input.ChangeSet)
	if err != nil {
		return err
	}

	if input.Source == "" {
		input.Source = SourceManual
	}
	if input.ActorType == "" {
		input.ActorType = "user"
	}
	if input.RequestID == "" {
		input.RequestID = middleware.RequestIDFromContext(ctx)
	}
	if principal, ok := platformauth.PrincipalFromContext(ctx); ok {
		if input.ActorID == nil {
			actorID := principal.UserID
			input.ActorID = &actorID
		}
		if input.SessionID == nil {
			sessionID := principal.SessionID
			input.SessionID = &sessionID
		}
	}

	var requestID *string
	if input.RequestID != "" {
		requestID = &input.RequestID
	}

	event := Event{
		ID:         uuid.New(),
		UserID:     input.UserID,
		Action:     input.Action,
		EntityType: input.EntityType,
		EntityID:   input.EntityID,
		Meta:       payload,
		Source:     input.Source,
		RequestID:  requestID,
		SessionID:  input.SessionID,
		ChangeSet:  changeSet,
		ActorType:  input.ActorType,
		ActorID:    input.ActorID,
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
			"meta":        input.Meta,
			"source":      event.Source,
			"request_id":  event.RequestID,
			"session_id":  event.SessionID,
			"change_set":  input.ChangeSet,
			"actor_type":  event.ActorType,
			"actor_id":    event.ActorID,
			"created_at":  event.CreatedAt,
		}
		if err := s.publisher.PublishJSON(ctx, s.topic, event.UserID.String(), message); err != nil && s.logger != nil {
			s.logger.Error("publish audit event", "error", err, "action", input.Action, "entity_type", input.EntityType)
		}
	}

	return nil
}
