package categories

import "github.com/google/uuid"

type CreateCategoryRequest struct {
	Kind     Kind       `json:"kind" validate:"required"`
	Name     string     `json:"name" validate:"required"`
	Color    *string    `json:"color"`
	Icon     *string    `json:"icon"`
	ParentID *uuid.UUID `json:"parent_id"`
}

type UpdateCategoryRequest struct {
	Name       *string    `json:"name"`
	Color      *string    `json:"color"`
	Icon       *string    `json:"icon"`
	ParentID   *uuid.UUID `json:"parent_id"`
	IsArchived *bool      `json:"is_archived"`
}
