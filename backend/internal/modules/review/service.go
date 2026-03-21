package review

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"moneyapp/backend/internal/core/audit"
	"moneyapp/backend/internal/core/common"
	"moneyapp/backend/internal/modules/finance/accounts"
	"moneyapp/backend/internal/modules/finance/transactions"
	"moneyapp/backend/internal/platform/clock"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/google/uuid"
)

type Service struct {
	repo         *Repository
	accountsRepo *accounts.Repository
	txRepo       *transactions.Repository
	audit        *audit.Service
	clock        clock.Clock
}

func NewService(repo *Repository, accountsRepo *accounts.Repository, txRepo *transactions.Repository, auditService *audit.Service, clock clock.Clock) *Service {
	return &Service{
		repo:         repo,
		accountsRepo: accountsRepo,
		txRepo:       txRepo,
		audit:        auditService,
		clock:        clock,
	}
}

func (s *Service) GetCurrent(ctx context.Context, userID uuid.UUID) (WeeklyReview, error) {
	periodStart, periodEnd := currentWeekBounds(s.clock.Now())
	review, err := s.repo.FindByPeriod(ctx, userID, periodStart.Format("2006-01-02"), periodEnd.Format("2006-01-02"))
	if err == nil {
		return review, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return WeeklyReview{}, err
	}

	expected, err := s.expectedBalance(ctx, userID, periodStart, periodEnd)
	if err != nil {
		return WeeklyReview{}, err
	}

	review = WeeklyReview{
		ID:              uuid.New(),
		UserID:          userID,
		PeriodStart:     periodStart,
		PeriodEnd:       periodEnd,
		ExpectedBalance: expected,
		Status:          StatusPending,
		CreatedAt:       s.clock.Now(),
	}
	if err := s.repo.Create(ctx, review); err != nil {
		return WeeklyReview{}, err
	}
	return review, nil
}

func (s *Service) SubmitBalance(ctx context.Context, userID, reviewID uuid.UUID, actual common.Money) (WeeklyReview, error) {
	review, err := s.repo.GetByID(ctx, userID, reviewID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WeeklyReview{}, httpx.NotFound("review_not_found", "review not found")
		}
		return WeeklyReview{}, err
	}

	delta := actual.Sub(review.ExpectedBalance)
	review.ActualBalance = &actual
	review.Delta = &delta
	if delta.IsZero() {
		review.Status = StatusMatched
	} else {
		review.Status = StatusDiscrepancyFound
	}
	completedAt := s.clock.Now()
	review.CompletedAt = &completedAt

	if err := s.repo.Update(ctx, review); err != nil {
		return WeeklyReview{}, err
	}

	if err := s.audit.Record(ctx, userID, "reviews.submit_balance", "weekly_review", &review.ID, map[string]any{
		"delta": delta.StringFixedBank(2),
	}); err != nil {
		return WeeklyReview{}, err
	}

	return review, nil
}

func (s *Service) Resolve(ctx context.Context, userID, reviewID uuid.UUID, note *string) (WeeklyReview, error) {
	review, err := s.repo.GetByID(ctx, userID, reviewID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WeeklyReview{}, httpx.NotFound("review_not_found", "review not found")
		}
		return WeeklyReview{}, err
	}
	review.Status = StatusResolved
	review.ResolutionNote = note
	completedAt := s.clock.Now()
	review.CompletedAt = &completedAt

	if err := s.repo.Update(ctx, review); err != nil {
		return WeeklyReview{}, err
	}
	if err := s.audit.Record(ctx, userID, "reviews.resolve", "weekly_review", &review.ID, map[string]any{}); err != nil {
		return WeeklyReview{}, err
	}
	return review, nil
}

func (s *Service) Skip(ctx context.Context, userID, reviewID uuid.UUID) (WeeklyReview, error) {
	review, err := s.repo.GetByID(ctx, userID, reviewID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WeeklyReview{}, httpx.NotFound("review_not_found", "review not found")
		}
		return WeeklyReview{}, err
	}
	review.Status = StatusSkipped
	completedAt := s.clock.Now()
	review.CompletedAt = &completedAt

	if err := s.repo.Update(ctx, review); err != nil {
		return WeeklyReview{}, err
	}
	if err := s.audit.Record(ctx, userID, "reviews.skip", "weekly_review", &review.ID, map[string]any{}); err != nil {
		return WeeklyReview{}, err
	}
	return review, nil
}

func (s *Service) expectedBalance(ctx context.Context, userID uuid.UUID, periodStart, periodEnd time.Time) (common.Money, error) {
	openingBalance, err := s.accountsRepo.TotalOpeningBalance(ctx, userID)
	if err != nil {
		return common.ZeroMoney(), err
	}
	netBefore, err := s.txRepo.NetEffectForUser(ctx, userID, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), periodStart)
	if err != nil {
		return common.ZeroMoney(), err
	}
	netDuring, err := s.txRepo.NetEffectForUser(ctx, userID, periodStart, periodEnd.AddDate(0, 0, 1))
	if err != nil {
		return common.ZeroMoney(), err
	}

	return openingBalance.Add(netBefore).Add(netDuring), nil
}

func currentWeekBounds(now time.Time) (time.Time, time.Time) {
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -(weekday - 1))
	end := start.AddDate(0, 0, 6)
	return start, end
}
