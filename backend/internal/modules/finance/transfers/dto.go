package transfers

import (
	"time"

	"moneyapp/backend/internal/core/common"

	"github.com/google/uuid"
)

type CreateTransferRequest struct {
	FromAccountID uuid.UUID    `json:"from_account_id" validate:"required"`
	ToAccountID   uuid.UUID    `json:"to_account_id" validate:"required"`
	Amount        common.Money `json:"amount"`
	Currency      string       `json:"currency" validate:"required,len=3"`
	Title         *string      `json:"title"`
	Note          *string      `json:"note"`
	OccurredAt    time.Time    `json:"occurred_at"`
}

type UpdateTransferRequest struct {
	FromAccountID *uuid.UUID    `json:"from_account_id"`
	ToAccountID   *uuid.UUID    `json:"to_account_id"`
	Amount        *common.Money `json:"amount"`
	Currency      *string       `json:"currency"`
	Title         *string       `json:"title"`
	Note          *string       `json:"note"`
	OccurredAt    *time.Time    `json:"occurred_at"`
}
