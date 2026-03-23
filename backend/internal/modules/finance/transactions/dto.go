package transactions

import (
	"time"

	"moneyapp/backend/internal/core/common"

	"github.com/google/uuid"
)

type CreateTransactionRequest struct {
	AccountID         uuid.UUID    `json:"account_id" validate:"required"`
	TransferAccountID *uuid.UUID   `json:"transfer_account_id"`
	Type              Type         `json:"type" validate:"required"`
	CategoryID        *uuid.UUID   `json:"category_id"`
	Amount            common.Money `json:"amount"`
	Currency          string       `json:"currency" validate:"required,len=3"`
	Direction         *Direction   `json:"direction"`
	Title             *string      `json:"title"`
	Note              *string      `json:"note"`
	IsMandatory       bool         `json:"is_mandatory"`
	IsSubscription    bool         `json:"is_subscription"`
	OccurredAt        time.Time    `json:"occurred_at"`
}

type UpdateTransactionRequest struct {
	AccountID         *uuid.UUID    `json:"account_id"`
	TransferAccountID *uuid.UUID    `json:"transfer_account_id"`
	Type              *Type         `json:"type"`
	CategoryID        *uuid.UUID    `json:"category_id"`
	Amount            *common.Money `json:"amount"`
	Currency          *string       `json:"currency"`
	Direction         *Direction    `json:"direction"`
	Title             *string       `json:"title"`
	Note              *string       `json:"note"`
	IsMandatory       *bool         `json:"is_mandatory"`
	IsSubscription    *bool         `json:"is_subscription"`
	OccurredAt        *time.Time    `json:"occurred_at"`
}

type ListFilters struct {
	AccountID        *uuid.UUID
	CategoryID       *uuid.UUID
	Type             *Type
	PostingState     *PostingState
	Source           *Source
	LinkedEntityType *string
	LinkedEntityID   *uuid.UUID
	DateFrom         *time.Time
	DateTo           *time.Time
	IncludeDeleted   bool
	Pagination       common.Pagination
}
