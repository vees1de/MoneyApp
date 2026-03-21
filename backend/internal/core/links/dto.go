package links

import "github.com/google/uuid"

type CreateLinkRequest struct {
	SourceType string         `json:"source_type" validate:"required"`
	SourceID   uuid.UUID      `json:"source_id" validate:"required"`
	TargetType string         `json:"target_type" validate:"required"`
	TargetID   uuid.UUID      `json:"target_id" validate:"required"`
	Relation   string         `json:"relation" validate:"required"`
	Meta       map[string]any `json:"meta"`
}

type ListByEntityQuery struct {
	EntityType string    `json:"entity_type" validate:"required"`
	EntityID   uuid.UUID `json:"entity_id" validate:"required"`
}
