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
			amount, currency, direction, posting_state, source, source_ref_id,
			template_id, recurring_rule_id, planned_expense_id, title, title_normalized, note,
			is_mandatory, is_subscription, base_currency, base_amount, fx_rate,
			occurred_at, created_at, updated_at
		)
		values (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17, $18,
			$19, $20, $21, $22, $23,
			$24, $25, $26
		)
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
		transaction.PostingState,
		transaction.Source,
		transaction.SourceRefID,
		transaction.TemplateID,
		transaction.RecurringRuleID,
		transaction.PlannedExpenseID,
		transaction.Title,
		transaction.TitleNormalized,
		transaction.Note,
		transaction.IsMandatory,
		transaction.IsSubscription,
		transaction.BaseCurrency,
		transaction.BaseAmount,
		transaction.FXRate,
		transaction.OccurredAt,
		transaction.CreatedAt,
		transaction.UpdatedAt,
	)
	return err
}

func (r *Repository) GetByID(ctx context.Context, userID, transactionID uuid.UUID, exec ...db.DBTX) (Transaction, error) {
	return r.getByID(ctx, userID, transactionID, false, exec...)
}

func (r *Repository) GetByIDIncludingDeleted(ctx context.Context, userID, transactionID uuid.UUID, exec ...db.DBTX) (Transaction, error) {
	return r.getByID(ctx, userID, transactionID, true, exec...)
}

func (r *Repository) getByID(ctx context.Context, userID, transactionID uuid.UUID, includeDeleted bool, exec ...db.DBTX) (Transaction, error) {
	query := `
		select id, user_id, account_id, transfer_account_id, type, category_id,
		       amount, currency, direction, posting_state, source, source_ref_id,
		       template_id, recurring_rule_id, planned_expense_id, title, title_normalized, note,
		       is_mandatory, is_subscription, base_currency, base_amount, fx_rate,
		       deleted_at, deleted_by, delete_reason, restored_at, occurred_at, created_at, updated_at
		from finance_transactions
		where id = $1 and user_id = $2
	`
	if !includeDeleted {
		query += ` and deleted_at is null`
	}

	var transaction Transaction
	err := scanTransactionRow(r.base(exec...).QueryRowContext(ctx, query, transactionID, userID), &transaction)
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
		    title_normalized = $11,
		    note = $12,
		    is_mandatory = $13,
		    is_subscription = $14,
		    occurred_at = $15,
		    updated_at = $16
		where id = $1 and user_id = $2 and deleted_at is null
	`
	result, err := r.base(exec...).ExecContext(ctx, query,
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
		transaction.TitleNormalized,
		transaction.Note,
		transaction.IsMandatory,
		transaction.IsSubscription,
		transaction.OccurredAt,
		transaction.UpdatedAt,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) SoftDelete(ctx context.Context, userID, transactionID uuid.UUID, deletedAt time.Time, deletedBy *uuid.UUID, reason *string, exec ...db.DBTX) error {
	query := `
		update finance_transactions
		set deleted_at = $3,
		    deleted_by = $4,
		    delete_reason = $5,
		    updated_at = $3
		where id = $1 and user_id = $2 and deleted_at is null
	`
	result, err := r.base(exec...).ExecContext(ctx, query, transactionID, userID, deletedAt, deletedBy, reason)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) Restore(ctx context.Context, userID, transactionID uuid.UUID, restoredAt time.Time, exec ...db.DBTX) error {
	query := `
		update finance_transactions
		set deleted_at = null,
		    deleted_by = null,
		    delete_reason = null,
		    restored_at = $3,
		    updated_at = $3
		where id = $1 and user_id = $2 and deleted_at is not null
	`
	result, err := r.base(exec...).ExecContext(ctx, query, transactionID, userID, restoredAt)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) List(ctx context.Context, userID uuid.UUID, filters ListFilters, exec ...db.DBTX) ([]Transaction, error) {
	var builder strings.Builder
	builder.WriteString(`
		select id, user_id, account_id, transfer_account_id, type, category_id,
		       amount, currency, direction, posting_state, source, source_ref_id,
		       template_id, recurring_rule_id, planned_expense_id, title, title_normalized, note,
		       is_mandatory, is_subscription, base_currency, base_amount, fx_rate,
		       deleted_at, deleted_by, delete_reason, restored_at, occurred_at, created_at, updated_at
		from finance_transactions t
		where t.user_id = $1
	`)

	args := []any{userID}
	index := 2

	if !filters.IncludeDeleted {
		builder.WriteString(" and t.deleted_at is null")
	}
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
	if filters.PostingState != nil {
		builder.WriteString(fmt.Sprintf(" and t.posting_state = $%d", index))
		args = append(args, *filters.PostingState)
		index++
	}
	if filters.Source != nil {
		builder.WriteString(fmt.Sprintf(" and t.source = $%d", index))
		args = append(args, *filters.Source)
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
		where user_id = $1
		  and posting_state = 'posted'
		  and deleted_at is null
		  and occurred_at >= $2
		  and occurred_at < $3
	`
	var total common.Money
	err := r.base(exec...).QueryRowContext(ctx, query, userID, start, end).Scan(&total)
	return total, err
}

func (r *Repository) SumByTypeForUser(ctx context.Context, userID uuid.UUID, txType Type, start, end time.Time, exec ...db.DBTX) (common.Money, error) {
	query := `
		select coalesce(sum(amount), 0)
		from finance_transactions
		where user_id = $1
		  and type = $2
		  and posting_state = 'posted'
		  and deleted_at is null
		  and occurred_at >= $3
		  and occurred_at < $4
	`
	var total common.Money
	err := r.base(exec...).QueryRowContext(ctx, query, userID, txType, start, end).Scan(&total)
	return total, err
}

func (r *Repository) ComputedAccountBalance(ctx context.Context, userID, accountID uuid.UUID, exec ...db.DBTX) (common.Money, error) {
	query := `
		select a.opening_balance + coalesce(sum(
			case
				when t.account_id = a.id and t.type = 'income' then t.amount
				when t.account_id = a.id and t.type = 'expense' then -t.amount
				when t.account_id = a.id and t.type = 'correction' and t.direction = 'inflow' then t.amount
				when t.account_id = a.id and t.type = 'correction' and t.direction = 'outflow' then -t.amount
				when t.account_id = a.id and t.type = 'transfer' then -t.amount
				when t.transfer_account_id = a.id and t.type = 'transfer' then t.amount
				else 0
			end
		), 0)
		from accounts a
		left join finance_transactions t
		  on t.user_id = a.user_id
		 and (t.account_id = a.id or t.transfer_account_id = a.id)
		 and t.posting_state = 'posted'
		 and t.deleted_at is null
		where a.user_id = $1 and a.id = $2
		group by a.opening_balance
	`
	var balance common.Money
	err := r.base(exec...).QueryRowContext(ctx, query, userID, accountID).Scan(&balance)
	return balance, err
}

type txScanner interface {
	Scan(dest ...any) error
}

func scanTransaction(scanner txScanner) (Transaction, error) {
	var item Transaction
	err := scanTransactionRow(scanner, &item)
	return item, err
}

func scanTransactionRow(scanner txScanner, item *Transaction) error {
	var titleNormalized sql.NullString
	var baseCurrency sql.NullString
	var baseAmount sql.NullString
	var fxRate sql.NullString

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
		&item.PostingState,
		&item.Source,
		&item.SourceRefID,
		&item.TemplateID,
		&item.RecurringRuleID,
		&item.PlannedExpenseID,
		&item.Title,
		&titleNormalized,
		&item.Note,
		&item.IsMandatory,
		&item.IsSubscription,
		&baseCurrency,
		&baseAmount,
		&fxRate,
		&item.DeletedAt,
		&item.DeletedBy,
		&item.DeleteReason,
		&item.RestoredAt,
		&item.OccurredAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if titleNormalized.Valid {
		item.TitleNormalized = &titleNormalized.String
	}
	if baseCurrency.Valid {
		item.BaseCurrency = &baseCurrency.String
	}
	if baseAmount.Valid {
		parsed, err := common.NewMoney(baseAmount.String)
		if err != nil {
			return err
		}
		item.BaseAmount = &parsed
	}
	if fxRate.Valid {
		item.FXRate = &fxRate.String
	}
	return nil
}
