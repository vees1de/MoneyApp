package links

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type EntityLink struct {
	ID         uuid.UUID       `json:"id"`
	UserID     uuid.UUID       `json:"user_id"`
	SourceType string          `json:"source_type"`
	SourceID   uuid.UUID       `json:"source_id"`
	TargetType string          `json:"target_type"`
	TargetID   uuid.UUID       `json:"target_id"`
	Relation   string          `json:"relation"`
	Meta       json.RawMessage `json:"meta,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}
