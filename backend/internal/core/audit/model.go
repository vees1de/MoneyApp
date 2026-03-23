package audit

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID         uuid.UUID       `json:"id"`
	UserID     uuid.UUID       `json:"user_id"`
	Action     string          `json:"action"`
	EntityType string          `json:"entity_type"`
	EntityID   *uuid.UUID      `json:"entity_id,omitempty"`
	Meta       json.RawMessage `json:"meta,omitempty"`
	Source     string          `json:"source"`
	RequestID  *string         `json:"request_id,omitempty"`
	SessionID  *uuid.UUID      `json:"session_id,omitempty"`
	ChangeSet  json.RawMessage `json:"change_set,omitempty"`
	ActorType  string          `json:"actor_type"`
	ActorID    *uuid.UUID      `json:"actor_id,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

const (
	SourceManual    = "manual"
	SourceRecurring = "recurring"
	SourceReview    = "review"
	SourceSystem    = "system"
)

type RecordInput struct {
	UserID     uuid.UUID
	Action     string
	EntityType string
	EntityID   *uuid.UUID
	Meta       map[string]any
	Source     string
	RequestID  string
	SessionID  *uuid.UUID
	ChangeSet  map[string]any
	ActorType  string
	ActorID    *uuid.UUID
}
