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

type PostingState string

const (
	PostingStateDraft  PostingState = "draft"
	PostingStatePosted PostingState = "posted"
)

type Source string

const (
	SourceManual    Source = "manual"
	SourceRecurring Source = "recurring"
	SourceReview    Source = "review"
	SourceSystem    Source = "system"
)

type Transaction struct {
	ID                uuid.UUID     `json:"id"`
	UserID            uuid.UUID     `json:"user_id"`
	AccountID         uuid.UUID     `json:"account_id"`
	TransferAccountID *uuid.UUID    `json:"transfer_account_id,omitempty"`
	Type              Type          `json:"type"`
	CategoryID        *uuid.UUID    `json:"category_id,omitempty"`
	Amount            common.Money  `json:"amount"`
	Currency          string        `json:"currency"`
	Direction         Direction     `json:"direction"`
	PostingState      PostingState  `json:"posting_state"`
	Source            Source        `json:"source"`
	SourceRefID       *uuid.UUID    `json:"source_ref_id,omitempty"`
	TemplateID        *uuid.UUID    `json:"template_id,omitempty"`
	RecurringRuleID   *uuid.UUID    `json:"recurring_rule_id,omitempty"`
	PlannedExpenseID  *uuid.UUID    `json:"planned_expense_id,omitempty"`
	Title             *string       `json:"title,omitempty"`
	TitleNormalized   *string       `json:"title_normalized,omitempty"`
	Note              *string       `json:"note,omitempty"`
	IsMandatory       bool          `json:"is_mandatory"`
	IsSubscription    bool          `json:"is_subscription"`
	BaseCurrency      *string       `json:"base_currency,omitempty"`
	BaseAmount        *common.Money `json:"base_amount,omitempty"`
	FXRate            *string       `json:"fx_rate,omitempty"`
	DeletedAt         *time.Time    `json:"deleted_at,omitempty"`
	DeletedBy         *uuid.UUID    `json:"deleted_by,omitempty"`
	DeleteReason      *string       `json:"delete_reason,omitempty"`
	RestoredAt        *time.Time    `json:"restored_at,omitempty"`
	OccurredAt        time.Time     `json:"occurred_at"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}
