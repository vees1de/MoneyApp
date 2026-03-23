package accounts

import (
	"context"
	"database/sql"
	"errors"

	"moneyapp/backend/internal/core/audit"
	"moneyapp/backend/internal/platform/clock"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/google/uuid"
)

type Service struct {
	db    *sql.DB
	repo  *Repository
	audit *audit.Service
	clock clock.Clock
}

func NewService(database *sql.DB, repo *Repository, auditService *audit.Service, clock clock.Clock) *Service {
	return &Service{
		db:    database,
		repo:  repo,
		audit: auditService,
		clock: clock,
	}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]Account, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) Get(ctx context.Context, userID, accountID uuid.UUID) (Account, error) {
	account, err := s.repo.GetByID(ctx, userID, accountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Account{}, httpx.NotFound("account_not_found", "account not found")
		}
		return Account{}, err
	}
	return account, nil
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, request CreateAccountRequest) (Account, error) {
	now := s.clock.Now()
	account := Account{
		ID:                 uuid.New(),
		UserID:             userID,
		Name:               request.Name,
		Kind:               request.Kind,
		Currency:           request.Currency,
		OpeningBalance:     request.OpeningBalance,
		CurrentBalance:     request.OpeningBalance,
		LastRecalculatedAt: &now,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := s.repo.Create(ctx, account); err != nil {
		return Account{}, err
	}

	if err := s.audit.RecordChange(ctx, audit.RecordInput{
		UserID:     userID,
		Action:     "accounts.create",
		EntityType: "account",
		EntityID:   &account.ID,
		Meta: map[string]any{
			"name": account.Name,
			"kind": account.Kind,
		},
		ChangeSet: map[string]any{
			"after": map[string]any{
				"name":        account.Name,
				"kind":        account.Kind,
				"currency":    account.Currency,
				"is_archived": account.IsArchived,
			},
		},
	}); err != nil {
		return Account{}, err
	}

	return account, nil
}

func (s *Service) Update(ctx context.Context, userID, accountID uuid.UUID, request UpdateAccountRequest) (Account, error) {
	account, err := s.Get(ctx, userID, accountID)
	if err != nil {
		return Account{}, err
	}

	if request.Name != nil {
		account.Name = *request.Name
	}
	if request.Kind != nil {
		account.Kind = *request.Kind
	}
	if request.IsArchived != nil {
		account.IsArchived = *request.IsArchived
	}
	account.UpdatedAt = s.clock.Now()

	if err := s.repo.Update(ctx, account); err != nil {
		return Account{}, err
	}

	if err := s.audit.RecordChange(ctx, audit.RecordInput{
		UserID:     userID,
		Action:     "accounts.update",
		EntityType: "account",
		EntityID:   &account.ID,
		Meta: map[string]any{
			"is_archived": account.IsArchived,
		},
		ChangeSet: map[string]any{
			"after": map[string]any{
				"name":        account.Name,
				"kind":        account.Kind,
				"is_archived": account.IsArchived,
			},
		},
	}); err != nil {
		return Account{}, err
	}

	return account, nil
}
