package transactions

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"moneyapp/backend/internal/core/audit"
	"moneyapp/backend/internal/modules/finance/accounts"
	"moneyapp/backend/internal/modules/finance/categories"
	"moneyapp/backend/internal/modules/finance/ledger"
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

func (s *Service) GetIncludingDeleted(ctx context.Context, userID, transactionID uuid.UUID) (Transaction, error) {
	item, err := s.repo.GetByIDIncludingDeleted(ctx, userID, transactionID)
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
		Currency:          strings.ToUpper(request.Currency),
		Direction:         directionForCreate(request),
		PostingState:      PostingStatePosted,
		Source:            SourceManual,
		Title:             request.Title,
		TitleNormalized:   normalizeTitle(request.Title),
		Note:              request.Note,
		IsMandatory:       request.IsMandatory,
		IsSubscription:    request.IsSubscription,
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
		return s.recalculateAffectedAccounts(ctx, userID, tx, transaction)
	})
	if err != nil {
		return Transaction{}, err
	}

	if err := s.audit.RecordChange(ctx, audit.RecordInput{
		UserID:     userID,
		Action:     "transactions.create",
		EntityType: "transaction",
		EntityID:   &transaction.ID,
		Meta: map[string]any{
			"type":   transaction.Type,
			"amount": transaction.Amount.StringFixedBank(2),
			"source": transaction.Source,
		},
		ChangeSet: map[string]any{
			"after": transactionSnapshot(transaction),
		},
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
		updated.Currency = strings.ToUpper(*request.Currency)
	}
	if request.Direction != nil {
		updated.Direction = *request.Direction
	} else if request.Type != nil {
		updated.Direction = defaultDirection(updated.Type, nil)
	}
	if request.Title != nil {
		updated.Title = request.Title
		updated.TitleNormalized = normalizeTitle(request.Title)
	}
	if request.Note != nil {
		updated.Note = request.Note
	}
	if request.IsMandatory != nil {
		updated.IsMandatory = *request.IsMandatory
	}
	if request.IsSubscription != nil {
		updated.IsSubscription = *request.IsSubscription
	}
	if request.OccurredAt != nil {
		updated.OccurredAt = *request.OccurredAt
	}
	updated.UpdatedAt = s.clock.Now()

	if err := s.validateTransaction(ctx, userID, updated); err != nil {
		return Transaction{}, err
	}

	err = platformdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.repo.Update(ctx, updated, tx); err != nil {
			return err
		}
		return s.recalculateAffectedAccounts(ctx, userID, tx, current, updated)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Transaction{}, httpx.NotFound("transaction_not_found", "transaction not found")
		}
		return Transaction{}, err
	}

	if err := s.audit.RecordChange(ctx, audit.RecordInput{
		UserID:     userID,
		Action:     "transactions.update",
		EntityType: "transaction",
		EntityID:   &updated.ID,
		Meta: map[string]any{
			"type":   updated.Type,
			"amount": updated.Amount.StringFixedBank(2),
		},
		ChangeSet: map[string]any{
			"before": transactionSnapshot(current),
			"after":  transactionSnapshot(updated),
		},
	}); err != nil {
		return Transaction{}, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, userID, transactionID uuid.UUID, reason *string) error {
	current, err := s.Get(ctx, userID, transactionID)
	if err != nil {
		return err
	}

	now := s.clock.Now()
	deletedBy := userID
	err = platformdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.repo.SoftDelete(ctx, userID, transactionID, now, &deletedBy, reason, tx); err != nil {
			return err
		}
		return s.recalculateAffectedAccounts(ctx, userID, tx, current)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return httpx.NotFound("transaction_not_found", "transaction not found")
		}
		return err
	}

	after := transactionSnapshot(current)
	after["deleted_at"] = now
	if reason != nil {
		after["delete_reason"] = *reason
	}
	return s.audit.RecordChange(ctx, audit.RecordInput{
		UserID:     userID,
		Action:     "transactions.delete",
		EntityType: "transaction",
		EntityID:   &transactionID,
		Meta: map[string]any{
			"type": current.Type,
		},
		ChangeSet: map[string]any{
			"before": transactionSnapshot(current),
			"after":  after,
		},
	})
}

func (s *Service) Restore(ctx context.Context, userID, transactionID uuid.UUID) (Transaction, error) {
	current, err := s.repo.GetByIDIncludingDeleted(ctx, userID, transactionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Transaction{}, httpx.NotFound("transaction_not_found", "transaction not found")
		}
		return Transaction{}, err
	}
	if current.DeletedAt == nil {
		return Transaction{}, httpx.Conflict("transaction_not_deleted", "transaction is not deleted")
	}
	if err := s.validateTransaction(ctx, userID, current); err != nil {
		return Transaction{}, err
	}

	now := s.clock.Now()
	err = platformdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.repo.Restore(ctx, userID, transactionID, now, tx); err != nil {
			return err
		}
		return s.recalculateAffectedAccounts(ctx, userID, tx, current)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Transaction{}, httpx.NotFound("transaction_not_found", "transaction not found")
		}
		return Transaction{}, err
	}

	restored, err := s.repo.GetByID(ctx, userID, transactionID)
	if err != nil {
		return Transaction{}, err
	}

	if err := s.audit.RecordChange(ctx, audit.RecordInput{
		UserID:     userID,
		Action:     "transactions.restore",
		EntityType: "transaction",
		EntityID:   &transactionID,
		Meta: map[string]any{
			"type": restored.Type,
		},
		ChangeSet: map[string]any{
			"before": transactionSnapshot(current),
			"after":  transactionSnapshot(restored),
		},
	}); err != nil {
		return Transaction{}, err
	}

	return restored, nil
}

func (s *Service) validateTransaction(ctx context.Context, userID uuid.UUID, item Transaction) error {
	account, err := s.accountsRepo.GetByID(ctx, userID, item.AccountID)
	if err != nil {
		return httpx.BadRequest("account_not_found", "account not found")
	}
	if account.IsArchived {
		return httpx.BadRequest("account_archived", "account is archived")
	}
	if !strings.EqualFold(account.Currency, item.Currency) {
		return httpx.BadRequest("currency_mismatch", "transaction currency must match account currency")
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
		if destination.IsArchived {
			return httpx.BadRequest("transfer_account_archived", "transfer account is archived")
		}
		if !strings.EqualFold(destination.Currency, account.Currency) {
			return httpx.BadRequest("currency_mismatch", "transfer accounts must use the same currency")
		}
	}

	if item.Type == TypeTransfer && item.CategoryID != nil {
		return httpx.BadRequest("transfer_category_not_allowed", "transfer cannot have a category")
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

func (s *Service) recalculateAffectedAccounts(ctx context.Context, userID uuid.UUID, tx *sql.Tx, items ...Transaction) error {
	entries := make([]ledger.Entry, 0, len(items))
	for _, item := range items {
		entries = append(entries, ledger.Entry{
			AccountID:         item.AccountID,
			TransferAccountID: item.TransferAccountID,
		})
	}

	recalculatedAt := s.clock.Now()
	for _, accountID := range ledger.AffectedAccountIDs(entries...) {
		balance, err := s.repo.ComputedAccountBalance(ctx, userID, accountID, tx)
		if err != nil {
			return err
		}
		if err := s.accountsRepo.SetCurrentBalance(ctx, userID, accountID, balance, recalculatedAt, tx); err != nil {
			return err
		}
	}
	return nil
}

func transactionSnapshot(item Transaction) map[string]any {
	snapshot := map[string]any{
		"account_id":      item.AccountID,
		"type":            item.Type,
		"amount":          item.Amount.StringFixedBank(2),
		"currency":        item.Currency,
		"direction":       item.Direction,
		"posting_state":   item.PostingState,
		"source":          item.Source,
		"is_mandatory":    item.IsMandatory,
		"is_subscription": item.IsSubscription,
		"occurred_at":     item.OccurredAt,
	}
	if item.TransferAccountID != nil {
		snapshot["transfer_account_id"] = *item.TransferAccountID
	}
	if item.CategoryID != nil {
		snapshot["category_id"] = *item.CategoryID
	}
	if item.Title != nil {
		snapshot["title"] = *item.Title
	}
	if item.Note != nil {
		snapshot["note"] = *item.Note
	}
	if item.DeletedAt != nil {
		snapshot["deleted_at"] = *item.DeletedAt
	}
	if item.DeleteReason != nil {
		snapshot["delete_reason"] = *item.DeleteReason
	}
	return snapshot
}

func normalizeOccurredAt(value, fallback time.Time) time.Time {
	if value.IsZero() {
		return fallback
	}
	return value
}

func normalizeTitle(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.TrimSpace(strings.ToLower(*value))
	if normalized == "" {
		return nil
	}
	return &normalized
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
