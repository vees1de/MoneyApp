package summary

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Monthly(ctx context.Context, userID uuid.UUID, now time.Time) (MonthlySummary, error) {
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)

	currentBalance, err := s.repo.CurrentBalance(ctx, userID)
	if err != nil {
		return MonthlySummary{}, err
	}
	income, err := s.repo.MonthlyIncome(ctx, userID, monthStart, monthEnd)
	if err != nil {
		return MonthlySummary{}, err
	}
	expense, err := s.repo.MonthlyExpense(ctx, userID, monthStart, monthEnd)
	if err != nil {
		return MonthlySummary{}, err
	}
	topCategories, err := s.repo.TopExpenseCategories(ctx, userID, monthStart, monthEnd, 3)
	if err != nil {
		return MonthlySummary{}, err
	}

	return MonthlySummary{
		CurrentBalance: currentBalance,
		IncomeTotal:    income,
		ExpenseTotal:   expense,
		TopCategories:  topCategories,
	}, nil
}
