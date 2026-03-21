package review

import (
	"time"

	"moneyapp/backend/internal/core/common"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending          Status = "pending"
	StatusMatched          Status = "matched"
	StatusDiscrepancyFound Status = "discrepancy_found"
	StatusResolved         Status = "resolved"
	StatusSkipped          Status = "skipped"
)

type WeeklyReview struct {
	ID              uuid.UUID     `json:"id"`
	UserID          uuid.UUID     `json:"user_id"`
	AccountID       *uuid.UUID    `json:"account_id,omitempty"`
	PeriodStart     time.Time     `json:"period_start"`
	PeriodEnd       time.Time     `json:"period_end"`
	ExpectedBalance common.Money  `json:"expected_balance"`
	ActualBalance   *common.Money `json:"actual_balance,omitempty"`
	Delta           *common.Money `json:"delta,omitempty"`
	Status          Status        `json:"status"`
	ResolutionNote  *string       `json:"resolution_note,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	CompletedAt     *time.Time    `json:"completed_at,omitempty"`
}
