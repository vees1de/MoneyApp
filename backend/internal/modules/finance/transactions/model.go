package transactions

import (
	"time"

	"moneyapp/backend/internal/core/common"

	"github.com/google/uuid"
)

type Type string

const (
	TypeIncome     Type = "income"
	TypeExpense    Type = "expense"
	TypeTransfer   Type = "transfer"
	TypeCorrection Type = "correction"
)

type Direction string

const (
	DirectionInflow   Direction = "inflow"
	DirectionOutflow  Direction = "outflow"
	DirectionInternal Direction = "internal"
)

type Transaction struct {
	ID                uuid.UUID    `json:"id"`
	UserID            uuid.UUID    `json:"user_id"`
	AccountID         uuid.UUID    `json:"account_id"`
	TransferAccountID *uuid.UUID   `json:"transfer_account_id,omitempty"`
	Type              Type         `json:"type"`
	CategoryID        *uuid.UUID   `json:"category_id,omitempty"`
	Amount            common.Money `json:"amount"`
	Currency          string       `json:"currency"`
	Direction         Direction    `json:"direction"`
	Title             *string      `json:"title,omitempty"`
	Note              *string      `json:"note,omitempty"`
	OccurredAt        time.Time    `json:"occurred_at"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
}
