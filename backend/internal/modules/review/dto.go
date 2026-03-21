package review

import "moneyapp/backend/internal/core/common"

type SubmitBalanceRequest struct {
	ActualBalance common.Money `json:"actual_balance"`
}

type ResolveRequest struct {
	ResolutionNote *string `json:"resolution_note"`
}
