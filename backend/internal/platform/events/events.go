package events

import (
	"context"
	"encoding/json"
	"time"

	"moneyapp/backend/internal/platform/db"

	"github.com/google/uuid"
)

type Message struct {
	ID         uuid.UUID
	Topic      string
	EventType  string
	EntityType string
	EntityID   uuid.UUID
	Payload    any
	OccurredAt time.Time
}

type Bus interface {
	Publish(context.Context, db.DBTX, Message) error
}

func MarshalPayload(payload any) ([]byte, error) {
	if payload == nil {
		return []byte("{}"), nil
	}

	return json.Marshal(payload)
}
