package savings

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"moneyapp/backend/internal/core/audit"
	"moneyapp/backend/internal/core/common"
	financesummary "moneyapp/backend/internal/modules/finance/summary"
	"moneyapp/backend/internal/platform/clock"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Service struct {
	repo        *Repository
	summaryRepo *financesummary.Repository
	audit       *audit.Service
	clock       clock.Clock
}

func NewService(repo *Repository, summaryRepo *financesummary.Repository, auditService *audit.Service, clock clock.Clock) *Service {
	return &Service{
		repo:        repo,
		summaryRepo: summaryRepo,
		audit:       auditService,
		clock:       clock,
	}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]Goal, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, request CreateGoalRequest) (Goal, error) {
	now := s.clock.Now()
	priority := request.Priority
	if priority == "" {
		priority = PriorityMedium
	}

	goal := Goal{
		ID:            uuid.New(),
		UserID:        userID,
		Title:         request.Title,
		TargetAmount:  request.TargetAmount,
		CurrentAmount: request.CurrentAmount,
		Currency:      request.Currency,
		TargetDate:    request.TargetDate,
		Priority:      priority,
		Status:        StatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repo.Create(ctx, goal); err != nil {
		return Goal{}, err
	}

	if err := s.audit.Record(ctx, userID, "savings.create", "savings_goal", &goal.ID, map[string]any{
		"title": goal.Title,
	}); err != nil {
		return Goal{}, err
	}

	return goal, nil
}

func (s *Service) Update(ctx context.Context, userID, goalID uuid.UUID, request UpdateGoalRequest) (Goal, error) {
	goal, err := s.repo.GetByID(ctx, userID, goalID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Goal{}, httpx.NotFound("goal_not_found", "goal not found")
		}
		return Goal{}, err
	}

	if request.Title != nil {
		goal.Title = *request.Title
	}
	if request.TargetAmount != nil {
		goal.TargetAmount = *request.TargetAmount
	}
	if request.CurrentAmount != nil {
		goal.CurrentAmount = *request.CurrentAmount
	}
	if request.TargetDate != nil {
		goal.TargetDate = request.TargetDate
	}
	if request.Priority != nil {
		goal.Priority = *request.Priority
	}
	if request.Status != nil {
		goal.Status = *request.Status
	}
	goal.UpdatedAt = s.clock.Now()

	if err := s.repo.Update(ctx, goal); err != nil {
		return Goal{}, err
	}
	if err := s.audit.Record(ctx, userID, "savings.update", "savings_goal", &goal.ID, map[string]any{
		"status": goal.Status,
	}); err != nil {
		return Goal{}, err
	}

	return goal, nil
}

func (s *Service) Summary(ctx context.Context, userID uuid.UUID) (Summary, error) {
	now := s.clock.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)

	income, err := s.summaryRepo.MonthlyIncome(ctx, userID, monthStart, monthEnd)
	if err != nil {
		return Summary{}, err
	}
	expense, err := s.summaryRepo.MonthlyExpense(ctx, userID, monthStart, monthEnd)
	if err != nil {
		return Summary{}, err
	}

	goals, err := s.repo.ListActiveByUser(ctx, userID)
	if err != nil {
		return Summary{}, err
	}

	totalTarget := common.ZeroMoney()
	totalCurrent := common.ZeroMoney()
	reserved := common.ZeroMoney()
	progressItems := make([]GoalProgress, 0, len(goals))
	for _, goal := range goals {
		totalTarget = totalTarget.Add(goal.TargetAmount)
		totalCurrent = totalCurrent.Add(goal.CurrentAmount)

		contribution := recommendedContribution(goal, now)
		reserved = reserved.Add(contribution)

		progressItems = append(progressItems, GoalProgress{
			Goal:                           goal,
			ProgressPercent:                progressPercent(goal),
			RecommendedMonthlyContribution: contribution,
		})
	}

	return Summary{
		TotalTarget:       totalTarget,
		TotalCurrent:      totalCurrent,
		ReservedThisMonth: reserved,
		SafeToSpend:       income.Sub(expense).Sub(reserved),
		Goals:             progressItems,
	}, nil
}

func recommendedContribution(goal Goal, now time.Time) common.Money {
	remaining := goal.TargetAmount.Sub(goal.CurrentAmount)
	if remaining.Decimal.LessThanOrEqual(decimal.Zero) {
		return common.ZeroMoney()
	}
	if goal.TargetDate == nil {
		return common.Money{Decimal: remaining.Decimal.Mul(decimal.NewFromFloat(0.1)).Round(2)}
	}

	monthsRemaining := monthsUntil(now, *goal.TargetDate)
	if monthsRemaining < 1 {
		monthsRemaining = 1
	}

	return common.Money{Decimal: remaining.Decimal.DivRound(decimal.NewFromInt(int64(monthsRemaining)), 2)}
}

func progressPercent(goal Goal) string {
	if goal.TargetAmount.IsZero() {
		return "0"
	}
	value := goal.CurrentAmount.Decimal.Div(goal.TargetAmount.Decimal).Mul(decimal.NewFromInt(100)).Round(0)
	return value.String()
}

func monthsUntil(now time.Time, target time.Time) int {
	yearDelta := target.Year() - now.Year()
	monthDelta := int(target.Month()) - int(now.Month())
	total := yearDelta*12 + monthDelta
	if target.Day() > now.Day() {
		total++
	}
	return total
}
