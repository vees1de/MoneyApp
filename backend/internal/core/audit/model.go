package audit

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID          uuid.UUID       `json:"id"`
	ActorUserID *uuid.UUID      `json:"actor_user_id,omitempty"`
	UserID      uuid.UUID       `json:"user_id"`
	EntityType  string          `json:"entity_type"`
	EntityID    *uuid.UUID      `json:"entity_id,omitempty"`
	Action      string          `json:"action"`
	OldValues   json.RawMessage `json:"old_values,omitempty"`
	NewValues   json.RawMessage `json:"new_values,omitempty"`
	Meta        json.RawMessage `json:"meta,omitempty"`
	Source      string          `json:"source"`
	RequestID   *string         `json:"request_id,omitempty"`
	SessionID   *uuid.UUID      `json:"session_id,omitempty"`
	ChangeSet   json.RawMessage `json:"change_set,omitempty"`
	ActorType   string          `json:"actor_type"`
	ActorID     *uuid.UUID      `json:"actor_id,omitempty"`
	IP          *string         `json:"ip,omitempty"`
	UserAgent   *string         `json:"user_agent,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

const (
	SourceManual    = "manual"
	SourceRecurring = "recurring"
	SourceReview    = "review"
	SourceSystem    = "system"
)

type RecordInput struct {
	UserID      uuid.UUID
	Action      string
	EntityType  string
	EntityID    *uuid.UUID
	Meta        map[string]any
	Source      string
	RequestID   string
	SessionID   *uuid.UUID
	ChangeSet   map[string]any
	OldValues   map[string]any
	NewValues   map[string]any
	ActorType   string
	ActorID     *uuid.UUID
	ActorUserID *uuid.UUID
	IP          *string
	UserAgent   *string
}
