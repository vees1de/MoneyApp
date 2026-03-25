package audit

import (
	"context"
	"encoding/json"

	"moneyapp/backend/internal/middleware"
	platformauth "moneyapp/backend/internal/platform/auth"
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

	return s.repo.Create(ctx, event)
}
