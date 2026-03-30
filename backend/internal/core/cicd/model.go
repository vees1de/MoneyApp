package cicd

import (
	"time"

	"github.com/google/uuid"
)

type SmokeCheck struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	SessionID uuid.UUID `json:"session_id"`
	RequestID string    `json:"request_id"`
	Trigger   string    `json:"trigger"`
	CreatedAt time.Time `json:"created_at"`
}

type SmokeCheckResult struct {
	Check     SmokeCheck `json:"check"`
	TotalRuns int        `json:"total_runs"`
}
