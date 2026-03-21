package categories

import (
	"time"

	"github.com/google/uuid"
)

type Kind string

const (
	KindIncome  Kind = "income"
	KindExpense Kind = "expense"
)

type Category struct {
	ID         uuid.UUID  `json:"id"`
	UserID     *uuid.UUID `json:"user_id,omitempty"`
	Kind       Kind       `json:"kind"`
	Name       string     `json:"name"`
	Color      *string    `json:"color,omitempty"`
	Icon       *string    `json:"icon,omitempty"`
	ParentID   *uuid.UUID `json:"parent_id,omitempty"`
	IsSystem   bool       `json:"is_system"`
	IsArchived bool       `json:"is_archived"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
