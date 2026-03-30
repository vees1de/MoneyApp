package outbox

import (
	"context"
	"time"

	"moneyapp/backend/internal/platform/db"
	"moneyapp/backend/internal/platform/events"

	"github.com/google/uuid"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Publish(ctx context.Context, exec db.DBTX, message events.Message) error {
	payload, err := events.MarshalPayload(message.Payload)
	if err != nil {
		return err
	}

	if message.ID == uuid.Nil {
		message.ID = uuid.New()
	}
	if message.OccurredAt.IsZero() {
		message.OccurredAt = time.Now().UTC()
	}

	if _, err := exec.ExecContext(ctx, `
		insert into domain_events (id, event_type, entity_type, entity_id, payload, occurred_at)
		values ($1, $2, $3, $4, $5, $6)
	`, message.ID, message.EventType, message.EntityType, message.EntityID, payload, message.OccurredAt); err != nil {
		return err
	}

	_, err = exec.ExecContext(ctx, `
		insert into outbox_messages (
			id, topic, event_type, entity_type, entity_id, payload, status, available_at, created_at
		)
		values ($1, $2, $3, $4, $5, $6, 'pending', $7, $7)
	`, uuid.New(), message.Topic, message.EventType, message.EntityType, message.EntityID, payload, message.OccurredAt)
	return err
}
