package cicd

import (
	"context"

	requestid "moneyapp/backend/internal/middleware"
	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/clock"

	"github.com/google/uuid"
)

const smokeTriggerSettingsButton = "settings_button"

type Service struct {
	repo  *Repository
	clock clock.Clock
}

func NewService(repo *Repository, appClock clock.Clock) *Service {
	return &Service{
		repo:  repo,
		clock: appClock,
	}
}

func (s *Service) Run(ctx context.Context, principal platformauth.Principal) (SmokeCheckResult, error) {
	requestID := requestid.RequestIDFromContext(ctx)
	if requestID == "" {
		requestID = uuid.NewString()
	}

	check := SmokeCheck{
		ID:        uuid.New(),
		UserID:    principal.UserID,
		SessionID: principal.SessionID,
		RequestID: requestID,
		Trigger:   smokeTriggerSettingsButton,
		CreatedAt: s.clock.Now(),
	}

	if err := s.repo.Create(ctx, check); err != nil {
		return SmokeCheckResult{}, err
	}

	totalRuns, err := s.repo.CountByUser(ctx, principal.UserID)
	if err != nil {
		return SmokeCheckResult{}, err
	}

	return SmokeCheckResult{
		Check:     check,
		TotalRuns: totalRuns,
	}, nil
}
