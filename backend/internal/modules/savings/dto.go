package savings

import (
	"time"

	"moneyapp/backend/internal/core/common"
)

type CreateGoalRequest struct {
	Title         string       `json:"title" validate:"required"`
	TargetAmount  common.Money `json:"target_amount"`
	CurrentAmount common.Money `json:"current_amount"`
	Currency      string       `json:"currency" validate:"required,len=3"`
	TargetDate    *time.Time   `json:"target_date"`
	Priority      Priority     `json:"priority"`
}

type UpdateGoalRequest struct {
	Title         *string       `json:"title"`
	TargetAmount  *common.Money `json:"target_amount"`
	CurrentAmount *common.Money `json:"current_amount"`
	TargetDate    *time.Time    `json:"target_date"`
	Priority      *Priority     `json:"priority"`
	Status        *Status       `json:"status"`
}

type GoalProgress struct {
	Goal                           Goal         `json:"goal"`
	ProgressPercent                string       `json:"progress_percent"`
	RecommendedMonthlyContribution common.Money `json:"recommended_monthly_contribution"`
}

type Summary struct {
	TotalTarget       common.Money   `json:"total_target"`
	TotalCurrent      common.Money   `json:"total_current"`
	ReservedThisMonth common.Money   `json:"reserved_this_month"`
	SafeToSpend       common.Money   `json:"safe_to_spend"`
	Goals             []GoalProgress `json:"goals"`
}
