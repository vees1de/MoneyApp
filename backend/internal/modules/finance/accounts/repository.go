package accounts

import (
	"context"
	"database/sql"
	"time"

	"moneyapp/backend/internal/core/common"
	"moneyapp/backend/internal/platform/db"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(database *sql.DB) *Repository {
	return &Repository{db: database}
}

func (r *Repository) base(exec ...db.DBTX) db.DBTX {
	if len(exec) > 0 && exec[0] != nil {
		return exec[0]
	}

	return r.db
}

func (r *Repository) Create(ctx context.Context, account Account, exec ...db.DBTX) error {
	query := `
		insert into accounts (
			id, user_id, name, kind, currency, opening_balance, current_balance,
			is_archived, last_recalculated_at, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.base(exec...).ExecContext(ctx, query,
		account.ID,
		account.UserID,
		account.Name,
		account.Kind,
		account.Currency,
		account.OpeningBalance,
		account.CurrentBalance,
		account.IsArchived,
		account.LastRecalculatedAt,
		account.CreatedAt,
		account.UpdatedAt,
	)
	return err
}

func (r *Repository) GetByID(ctx context.Context, userID, accountID uuid.UUID, exec ...db.DBTX) (Account, error) {
	query := `
		select id, user_id, name, kind, currency, opening_balance, current_balance,
		       is_archived, last_recalculated_at, created_at, updated_at
		from accounts
		where id = $1 and user_id = $2
	`
	return r.scanOne(ctx, query, accountID, userID, exec...)
}

func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) ([]Account, error) {
	query := `
		select id, user_id, name, kind, currency, opening_balance, current_balance,
		       is_archived, last_recalculated_at, created_at, updated_at
		from accounts
		where user_id = $1
		order by created_at asc
	`
	rows, err := r.base(exec...).QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Account
	for rows.Next() {
		account, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, account)
	}

	return result, rows.Err()
}

func (r *Repository) ListActiveByUser(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) ([]Account, error) {
	query := `
		select id, user_id, name, kind, currency, opening_balance, current_balance,
		       is_archived, last_recalculated_at, created_at, updated_at
		from accounts
		where user_id = $1 and is_archived = false
		order by created_at asc
	`
	rows, err := r.base(exec...).QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Account
	for rows.Next() {
		account, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, account)
	}

	return result, rows.Err()
}

func (r *Repository) Update(ctx context.Context, account Account, exec ...db.DBTX) error {
	query := `
		update accounts
		set name = $3,
		    kind = $4,
		    is_archived = $5,
		    updated_at = $6
		where id = $1 and user_id = $2
	`
	_, err := r.base(exec...).ExecContext(ctx, query, account.ID, account.UserID, account.Name, account.Kind, account.IsArchived, account.UpdatedAt)
	return err
}

func (r *Repository) AdjustBalance(ctx context.Context, userID, accountID uuid.UUID, delta common.Money, exec ...db.DBTX) error {
	query := `
		update accounts
		set current_balance = current_balance + $3,
		    updated_at = now()
		where id = $1 and user_id = $2
	`
	_, err := r.base(exec...).ExecContext(ctx, query, accountID, userID, delta)
	return err
}

func (r *Repository) SetCurrentBalance(ctx context.Context, userID, accountID uuid.UUID, balance common.Money, recalculatedAt time.Time, exec ...db.DBTX) error {
	query := `
		update accounts
		set current_balance = $3,
		    last_recalculated_at = $4,
		    updated_at = $4
		where id = $1 and user_id = $2
	`
	_, err := r.base(exec...).ExecContext(ctx, query, accountID, userID, balance, recalculatedAt)
	return err
}

func (r *Repository) TotalCurrentBalance(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) (common.Money, error) {
	query := `
		select coalesce(sum(current_balance), 0)
		from accounts
		where user_id = $1 and is_archived = false
	`
	var total common.Money
	err := r.base(exec...).QueryRowContext(ctx, query, userID).Scan(&total)
	return total, err
}

func (r *Repository) TotalOpeningBalance(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) (common.Money, error) {
	query := `
		select coalesce(sum(opening_balance), 0)
		from accounts
		where user_id = $1 and is_archived = false
	`
	var total common.Money
	err := r.base(exec...).QueryRowContext(ctx, query, userID).Scan(&total)
	return total, err
}

func (r *Repository) scanOne(ctx context.Context, query string, arg1, arg2 any, exec ...db.DBTX) (Account, error) {
	row := r.base(exec...).QueryRowContext(ctx, query, arg1, arg2)
	var account Account
	err := row.Scan(
		&account.ID,
		&account.UserID,
		&account.Name,
		&account.Kind,
		&account.Currency,
		&account.OpeningBalance,
		&account.CurrentBalance,
		&account.IsArchived,
		&account.LastRecalculatedAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	return account, err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAccount(scanner rowScanner) (Account, error) {
	var account Account
	err := scanner.Scan(
		&account.ID,
		&account.UserID,
		&account.Name,
		&account.Kind,
		&account.Currency,
		&account.OpeningBalance,
		&account.CurrentBalance,
		&account.IsArchived,
		&account.LastRecalculatedAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	return account, err
}
