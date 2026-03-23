package accounts

import "moneyapp/backend/internal/core/common"

type CreateAccountRequest struct {
	Name           string       `json:"name" validate:"required"`
	Kind           Kind         `json:"kind" validate:"required"`
	Currency       string       `json:"currency" validate:"required,len=3"`
	OpeningBalance common.Money `json:"opening_balance"`
}

type UpdateAccountRequest struct {
	Name       *string `json:"name"`
	Kind       *Kind   `json:"kind"`
	IsArchived *bool   `json:"is_archived"`
}
