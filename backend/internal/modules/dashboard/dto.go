package dashboard

import (
	"moneyapp/backend/internal/core/common"
	financesummary "moneyapp/backend/internal/modules/finance/summary"
	"moneyapp/backend/internal/modules/review"
	"moneyapp/backend/internal/modules/savings"
)

type FinanceDashboard struct {
	CurrentBalance common.Money                 `json:"current_balance"`
	MonthlyIncome  common.Money                 `json:"monthly_income"`
	MonthlyExpense common.Money                 `json:"monthly_expense"`
	SavedThisMonth common.Money                 `json:"saved_this_month"`
	SafeToSpend    common.Money                 `json:"safe_to_spend"`
	TopCategories  []financesummary.TopCategory `json:"top_categories"`
	Savings        []savings.GoalProgress       `json:"savings"`
	WeeklyReview   review.WeeklyReview          `json:"weekly_review"`
	Insights       []string                     `json:"insights"`
}
