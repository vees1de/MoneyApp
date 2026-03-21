package summary

import "moneyapp/backend/internal/core/common"

type TopCategory struct {
	CategoryID *string      `json:"category_id,omitempty"`
	Name       string       `json:"name"`
	Amount     common.Money `json:"amount"`
}

type MonthlySummary struct {
	CurrentBalance common.Money  `json:"current_balance"`
	IncomeTotal    common.Money  `json:"income_total"`
	ExpenseTotal   common.Money  `json:"expense_total"`
	TopCategories  []TopCategory `json:"top_categories"`
}
