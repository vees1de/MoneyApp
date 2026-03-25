package dashboard

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"moneyapp/backend/internal/modules/finance/summary"
	"moneyapp/backend/internal/modules/review"
	"moneyapp/backend/internal/modules/savings"
	"moneyapp/backend/internal/platform/clock"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Service struct {
	repo           *Repository
	summaryService *summary.Service
	savingsService *savings.Service
	reviewService  *review.Service
	clock          clock.Clock
}

func NewService(repo *Repository, summaryService *summary.Service, savingsService *savings.Service, reviewService *review.Service, clock clock.Clock) *Service {
	return &Service{
		repo:           repo,
		summaryService: summaryService,
		savingsService: savingsService,
		reviewService:  reviewService,
		clock:          clock,
	}
}

func (s *Service) Finance(ctx context.Context, userID uuid.UUID) (FinanceDashboard, error) {
	now := s.clock.Now()
	monthly, err := s.summaryService.Monthly(ctx, userID, now)
	if err != nil {
		return FinanceDashboard{}, err
	}
	savingsSummary, err := s.savingsService.Summary(ctx, userID)
	if err != nil {
		return FinanceDashboard{}, err
	}
	currentReview, err := s.reviewService.GetCurrent(ctx, userID)
	if err != nil {
		return FinanceDashboard{}, err
	}

	insights := []string{}
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)
	trend, err := s.repo.TopExpenseTrend(ctx, userID, monthStart, monthEnd)
	if err == nil && trend.Average.Decimal.GreaterThan(decimal.Zero) {
		threshold := trend.Average.Decimal.Mul(decimal.NewFromFloat(1.3))
		if trend.Current.Decimal.GreaterThan(threshold) {
			increase := trend.Current.Decimal.Div(trend.Average.Decimal).Sub(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(100)).Round(0)
			insights = append(insights, fmt.Sprintf("По категории '%s' расходы выше среднего на %s%%", trend.Name, increase.String()))
		}
	}
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		return FinanceDashboard{}, err
	}

	if monthly.ExpenseTotal.Decimal.GreaterThan(monthly.IncomeTotal.Decimal) {
		insights = append(insights, "Расходы в текущем месяце уже выше доходов")
	}
	if savingsSummary.SafeToSpend.Decimal.LessThan(decimal.Zero) {
		insights = append(insights, "Safe-to-spend ушёл в минус: нужно сократить discretionary расходы")
	}

	return FinanceDashboard{
		CurrentBalance: monthly.CurrentBalance,
		MonthlyIncome:  monthly.IncomeTotal,
		MonthlyExpense: monthly.ExpenseTotal,
		SavedThisMonth: savingsSummary.ReservedThisMonth,
		SafeToSpend:    savingsSummary.SafeToSpend,
		TopCategories:  monthly.TopCategories,
		Savings:        savingsSummary.Goals,
		WeeklyReview:   currentReview,
		Insights:       insights,
	}, nil
}
