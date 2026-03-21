package transactions

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"moneyapp/backend/internal/core/audit"
	"moneyapp/backend/internal/core/common"
	"moneyapp/backend/internal/modules/finance/accounts"
	"moneyapp/backend/internal/modules/finance/categories"
	"moneyapp/backend/internal/platform/clock"
	platformdb "moneyapp/backend/internal/platform/db"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/google/uuid"
)

type Service struct {
	db           *sql.DB
	repo         *Repository
	accountsRepo *accounts.Repository
	categoryRepo *categories.Repository
	audit        *audit.Service
	clock        clock.Clock
}

func NewService(database *sql.DB, repo *Repository, accountsRepo *accounts.Repository, categoryRepo *categories.Repository, auditService *audit.Service, clock clock.Clock) *Service {
	return &Service{
		db:           database,
		repo:         repo,
		accountsRepo: accountsRepo,
		categoryRepo: categoryRepo,
		audit:        auditService,
		clock:        clock,
	}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID, filters ListFilters) ([]Transaction, error) {
	return s.repo.List(ctx, userID, filters)
}

func (s *Service) Get(ctx context.Context, userID, transactionID uuid.UUID) (Transaction, error) {
	item, err := s.repo.GetByID(ctx, userID, transactionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Transaction{}, httpx.NotFound("transaction_not_found", "transaction not found")
		}
		return Transaction{}, err
	}
	return item, nil
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, request CreateTransactionRequest) (Transaction, error) {
	now := s.clock.Now()
	transaction := Transaction{
		ID:                uuid.New(),
		UserID:            userID,
		AccountID:         request.AccountID,
		TransferAccountID: request.TransferAccountID,
		Type:              request.Type,
		CategoryID:        request.CategoryID,
		Amount:            request.Amount,
		Currency:          request.Currency,
		Direction:         directionForCreate(request),
		Title:             request.Title,
		Note:              request.Note,
		OccurredAt:        normalizeOccurredAt(request.OccurredAt, now),
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := s.validateTransaction(ctx, userID, transaction); err != nil {
		return Transaction{}, err
	}

	err := platformdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.repo.Create(ctx, transaction, tx); err != nil {
			return err
		}
		return s.applyBalanceEffects(ctx, transaction, tx)
	})
	if err != nil {
		return Transaction{}, err
	}

	if err := s.audit.Record(ctx, userID, "transactions.create", "transaction", &transaction.ID, map[string]any{
		"type":   transaction.Type,
		"amount": transaction.Amount.StringFixedBank(2),
	}); err != nil {
		return Transaction{}, err
	}

	return transaction, nil
}

func (s *Service) Update(ctx context.Context, userID, transactionID uuid.UUID, request UpdateTransactionRequest) (Transaction, error) {
	current, err := s.Get(ctx, userID, transactionID)
	if err != nil {
		return Transaction{}, err
	}

	updated := current
	if request.AccountID != nil {
		updated.AccountID = *request.AccountID
	}
	if request.TransferAccountID != nil {
		updated.TransferAccountID = request.TransferAccountID
	}
	if request.Type != nil {
		updated.Type = *request.Type
		if updated.Type != TypeTransfer {
			updated.TransferAccountID = nil
		}
	}
	if request.CategoryID != nil {
		updated.CategoryID = request.CategoryID
	}
	if request.Amount != nil {
		updated.Amount = *request.Amount
	}
	if request.Currency != nil {
		updated.Currency = *request.Currency
	}
	if request.Direction != nil {
		updated.Direction = *request.Direction
	} else if request.Type != nil {
		updated.Direction = defaultDirection(updated.Type, nil)
	}
	if request.Title != nil {
		updated.Title = request.Title
	}
	if request.Note != nil {
		updated.Note = request.Note
	}
	if request.OccurredAt != nil {
		updated.OccurredAt = *request.OccurredAt
	}
	updated.UpdatedAt = s.clock.Now()

	if err := s.validateTransaction(ctx, userID, updated); err != nil {
		return Transaction{}, err
	}

	err = platformdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.revertBalanceEffects(ctx, current, tx); err != nil {
			return err
		}
		if err := s.repo.Update(ctx, updated, tx); err != nil {
			return err
		}
		return s.applyBalanceEffects(ctx, updated, tx)
	})
	if err != nil {
		return Transaction{}, err
	}

	if err := s.audit.Record(ctx, userID, "transactions.update", "transaction", &updated.ID, map[string]any{
		"type":   updated.Type,
		"amount": updated.Amount.StringFixedBank(2),
	}); err != nil {
		return Transaction{}, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, userID, transactionID uuid.UUID) error {
	current, err := s.Get(ctx, userID, transactionID)
	if err != nil {
		return err
	}

	err = platformdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.revertBalanceEffects(ctx, current, tx); err != nil {
			return err
		}
		return s.repo.Delete(ctx, userID, transactionID, tx)
	})
	if err != nil {
		return err
	}

	return s.audit.Record(ctx, userID, "transactions.delete", "transaction", &transactionID, map[string]any{
		"type": current.Type,
	})
}

func (s *Service) validateTransaction(ctx context.Context, userID uuid.UUID, item Transaction) error {
	account, err := s.accountsRepo.GetByID(ctx, userID, item.AccountID)
	if err != nil {
		return httpx.BadRequest("account_not_found", "account not found")
	}
	if account.IsArchived {
		return httpx.BadRequest("account_archived", "account is archived")
	}

	if item.TransferAccountID != nil {
		if item.Type != TypeTransfer {
			return httpx.BadRequest("transfer_account_not_allowed", "transfer_account_id is allowed only for transfer transactions")
		}
		if *item.TransferAccountID == item.AccountID {
			return httpx.BadRequest("same_transfer_account", "transfer accounts must be different")
		}
		destination, err := s.accountsRepo.GetByID(ctx, userID, *item.TransferAccountID)
		if err != nil {
			return httpx.BadRequest("transfer_account_not_found", "transfer account not found")
		}
		if destination.Currency != account.Currency {
			return httpx.BadRequest("currency_mismatch", "transfer accounts must use the same currency")
		}
	}

	if item.CategoryID != nil {
		category, err := s.categoryRepo.GetByID(ctx, userID, *item.CategoryID)
		if err != nil {
			return httpx.BadRequest("category_not_found", "category not found")
		}
		if category.IsArchived {
			return httpx.BadRequest("category_archived", "category is archived")
		}
		if item.Type == TypeIncome && category.Kind != categories.KindIncome {
			return httpx.BadRequest("category_kind_mismatch", "income transaction requires income category")
		}
		if item.Type == TypeExpense && category.Kind != categories.KindExpense {
			return httpx.BadRequest("category_kind_mismatch", "expense transaction requires expense category")
		}
	}

	if item.Amount.IsZero() {
		return httpx.BadRequest("amount_required", "amount must be greater than zero")
	}

	if item.Type == TypeTransfer && item.TransferAccountID == nil {
		return httpx.BadRequest("transfer_account_required", "transfer_account_id is required for transfers")
	}

	if item.Type == TypeCorrection && item.Direction != DirectionInflow && item.Direction != DirectionOutflow {
		return httpx.BadRequest("invalid_correction_direction", "correction requires inflow or outflow direction")
	}

	return nil
}

func (s *Service) applyBalanceEffects(ctx context.Context, item Transaction, tx *sql.Tx) error {
	for _, effect := range balanceEffects(item) {
		if effect.Delta.IsZero() {
			continue
		}
		if err := s.accountsRepo.AdjustBalance(ctx, item.UserID, effect.AccountID, effect.Delta, tx); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) revertBalanceEffects(ctx context.Context, item Transaction, tx *sql.Tx) error {
	for _, effect := range balanceEffects(item) {
		if effect.Delta.IsZero() {
			continue
		}
		if err := s.accountsRepo.AdjustBalance(ctx, item.UserID, effect.AccountID, effect.Delta.Neg(), tx); err != nil {
			return err
		}
	}
	return nil
}

type balanceEffect struct {
	AccountID uuid.UUID
	Delta     common.Money
}

func balanceEffects(item Transaction) []balanceEffect {
	switch item.Type {
	case TypeIncome:
		return []balanceEffect{{AccountID: item.AccountID, Delta: item.Amount}}
	case TypeExpense:
		return []balanceEffect{{AccountID: item.AccountID, Delta: item.Amount.Neg()}}
	case TypeCorrection:
		delta := item.Amount
		if item.Direction == DirectionOutflow {
			delta = item.Amount.Neg()
		}
		return []balanceEffect{{AccountID: item.AccountID, Delta: delta}}
	case TypeTransfer:
		if item.TransferAccountID == nil {
			return nil
		}
		return []balanceEffect{
			{AccountID: item.AccountID, Delta: item.Amount.Neg()},
			{AccountID: *item.TransferAccountID, Delta: item.Amount},
		}
	default:
		return nil
	}
}

func normalizeOccurredAt(value, fallback time.Time) time.Time {
	if value.IsZero() {
		return fallback
	}
	return value
}

func directionForCreate(request CreateTransactionRequest) Direction {
	return defaultDirection(request.Type, request.Direction)
}

func defaultDirection(txType Type, requested *Direction) Direction {
	if requested != nil {
		return *requested
	}
	switch txType {
	case TypeIncome:
		return DirectionInflow
	case TypeExpense:
		return DirectionOutflow
	case TypeTransfer:
		return DirectionInternal
	default:
		return DirectionInflow
	}
}
