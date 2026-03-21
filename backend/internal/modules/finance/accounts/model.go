package accounts

import (
	"time"

	"moneyapp/backend/internal/core/common"

	"github.com/google/uuid"
)

type Kind string

const (
	KindCash        Kind = "cash"
	KindBankCard    Kind = "bank_card"
	KindBankAccount Kind = "bank_account"
	KindSavings     Kind = "savings"
	KindVirtual     Kind = "virtual"
)

type Account struct {
	ID             uuid.UUID    `json:"id"`
	UserID         uuid.UUID    `json:"user_id"`
	Name           string       `json:"name"`
	Kind           Kind         `json:"kind"`
	Currency       string       `json:"currency"`
	OpeningBalance common.Money `json:"opening_balance"`
	CurrentBalance common.Money `json:"current_balance"`
	IsArchived     bool         `json:"is_archived"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}
