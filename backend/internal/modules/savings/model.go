package savings

import (
	"time"

	"moneyapp/backend/internal/core/common"

	"github.com/google/uuid"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusPaused    Status = "paused"
	StatusCompleted Status = "completed"
	StatusArchived  Status = "archived"
)

type Goal struct {
	ID            uuid.UUID    `json:"id"`
	UserID        uuid.UUID    `json:"user_id"`
	Title         string       `json:"title"`
	TargetAmount  common.Money `json:"target_amount"`
	CurrentAmount common.Money `json:"current_amount"`
	Currency      string       `json:"currency"`
	TargetDate    *time.Time   `json:"target_date,omitempty"`
	Priority      Priority     `json:"priority"`
	Status        Status       `json:"status"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}
