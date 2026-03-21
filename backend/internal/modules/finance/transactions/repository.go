package transactions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

func (r *Repository) Create(ctx context.Context, transaction Transaction, exec ...db.DBTX) error {
	query := `
		insert into finance_transactions (
			id, user_id, account_id, transfer_account_id, type, category_id,
			amount, currency, direction, title, note, occurred_at, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := r.base(exec...).ExecContext(ctx, query,
		transaction.ID,
		transaction.UserID,
		transaction.AccountID,
		transaction.TransferAccountID,
		transaction.Type,
		transaction.CategoryID,
		transaction.Amount,
		transaction.Currency,
		transaction.Direction,
		transaction.Title,
		transaction.Note,
		transaction.OccurredAt,
		transaction.CreatedAt,
		transaction.UpdatedAt,
	)
	return err
}

func (r *Repository) GetByID(ctx context.Context, userID, transactionID uuid.UUID, exec ...db.DBTX) (Transaction, error) {
	query := `
		select id, user_id, account_id, transfer_account_id, type, category_id,
		       amount, currency, direction, title, note, occurred_at, created_at, updated_at
		from finance_transactions
		where id = $1 and user_id = $2
	`
	var transaction Transaction
	err := r.base(exec...).QueryRowContext(ctx, query, transactionID, userID).Scan(
		&transaction.ID,
		&transaction.UserID,
		&transaction.AccountID,
		&transaction.TransferAccountID,
		&transaction.Type,
		&transaction.CategoryID,
		&transaction.Amount,
		&transaction.Currency,
		&transaction.Direction,
		&transaction.Title,
		&transaction.Note,
		&transaction.OccurredAt,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)
	return transaction, err
}

func (r *Repository) Update(ctx context.Context, transaction Transaction, exec ...db.DBTX) error {
	query := `
		update finance_transactions
		set account_id = $3,
		    transfer_account_id = $4,
		    type = $5,
		    category_id = $6,
		    amount = $7,
		    currency = $8,
		    direction = $9,
		    title = $10,
		    note = $11,
		    occurred_at = $12,
		    updated_at = $13
		where id = $1 and user_id = $2
	`
	_, err := r.base(exec...).ExecContext(ctx, query,
		transaction.ID,
		transaction.UserID,
		transaction.AccountID,
		transaction.TransferAccountID,
		transaction.Type,
		transaction.CategoryID,
		transaction.Amount,
		transaction.Currency,
		transaction.Direction,
		transaction.Title,
		transaction.Note,
		transaction.OccurredAt,
		transaction.UpdatedAt,
	)
	return err
}

func (r *Repository) Delete(ctx context.Context, userID, transactionID uuid.UUID, exec ...db.DBTX) error {
	query := `delete from finance_transactions where id = $1 and user_id = $2`
	_, err := r.base(exec...).ExecContext(ctx, query, transactionID, userID)
	return err
}

func (r *Repository) List(ctx context.Context, userID uuid.UUID, filters ListFilters, exec ...db.DBTX) ([]Transaction, error) {
	var builder strings.Builder
	builder.WriteString(`
		select id, user_id, account_id, transfer_account_id, type, category_id,
		       amount, currency, direction, title, note, occurred_at, created_at, updated_at
		from finance_transactions t
		where t.user_id = $1
	`)

	args := []any{userID}
	index := 2
	if filters.AccountID != nil {
		builder.WriteString(fmt.Sprintf(" and (t.account_id = $%d or t.transfer_account_id = $%d)", index, index))
		args = append(args, *filters.AccountID)
		index++
	}
	if filters.CategoryID != nil {
		builder.WriteString(fmt.Sprintf(" and t.category_id = $%d", index))
		args = append(args, *filters.CategoryID)
		index++
	}
	if filters.Type != nil {
		builder.WriteString(fmt.Sprintf(" and t.type = $%d", index))
		args = append(args, *filters.Type)
		index++
	}
	if filters.DateFrom != nil {
		builder.WriteString(fmt.Sprintf(" and t.occurred_at >= $%d", index))
		args = append(args, *filters.DateFrom)
		index++
	}
	if filters.DateTo != nil {
		builder.WriteString(fmt.Sprintf(" and t.occurred_at <= $%d", index))
		args = append(args, *filters.DateTo)
		index++
	}
	if filters.LinkedEntityType != nil && filters.LinkedEntityID != nil {
		builder.WriteString(fmt.Sprintf(`
			and exists (
				select 1
				from entity_links el
				where el.user_id = t.user_id
				  and (
				       (el.source_type = 'transaction' and el.source_id = t.id and el.target_type = $%d and el.target_id = $%d)
				    or (el.target_type = 'transaction' and el.target_id = t.id and el.source_type = $%d and el.source_id = $%d)
				  )
			)
		`, index, index+1, index, index+1))
		args = append(args, *filters.LinkedEntityType, *filters.LinkedEntityID)
		index += 2
	}

	builder.WriteString(fmt.Sprintf(" order by t.occurred_at desc, t.created_at desc limit $%d offset $%d", index, index+1))
	args = append(args, filters.Pagination.Limit, filters.Pagination.Offset)

	rows, err := r.base(exec...).QueryContext(ctx, builder.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Transaction
	for rows.Next() {
		item, err := scanTransaction(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, rows.Err()
}

func (r *Repository) NetEffectForUser(ctx context.Context, userID uuid.UUID, start, end time.Time, exec ...db.DBTX) (common.Money, error) {
	query := `
		select coalesce(sum(
			case
				when type = 'income' then amount
				when type = 'expense' then -amount
				when type = 'correction' and direction = 'inflow' then amount
				when type = 'correction' and direction = 'outflow' then -amount
				else 0
			end
		), 0)
		from finance_transactions
		where user_id = $1 and occurred_at >= $2 and occurred_at < $3
	`
	var total common.Money
	err := r.base(exec...).QueryRowContext(ctx, query, userID, start, end).Scan(&total)
	return total, err
}

func (r *Repository) SumByTypeForUser(ctx context.Context, userID uuid.UUID, txType Type, start, end time.Time, exec ...db.DBTX) (common.Money, error) {
	query := `
		select coalesce(sum(amount), 0)
		from finance_transactions
		where user_id = $1 and type = $2 and occurred_at >= $3 and occurred_at < $4
	`
	var total common.Money
	err := r.base(exec...).QueryRowContext(ctx, query, userID, txType, start, end).Scan(&total)
	return total, err
}

type txScanner interface {
	Scan(dest ...any) error
}

func scanTransaction(scanner txScanner) (Transaction, error) {
	var item Transaction
	err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&item.AccountID,
		&item.TransferAccountID,
		&item.Type,
		&item.CategoryID,
		&item.Amount,
		&item.Currency,
		&item.Direction,
		&item.Title,
		&item.Note,
		&item.OccurredAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	return item, err
}
