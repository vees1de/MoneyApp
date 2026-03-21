package summary

import (
	"context"
	"database/sql"
	"time"

	"moneyapp/backend/internal/core/common"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(database *sql.DB) *Repository {
	return &Repository{db: database}
}

func (r *Repository) CurrentBalance(ctx context.Context, userID uuid.UUID) (common.Money, error) {
	query := `
		select coalesce(sum(current_balance), 0)
		from accounts
		where user_id = $1 and is_archived = false
	`
	var total common.Money
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&total)
	return total, err
}

func (r *Repository) MonthlyIncome(ctx context.Context, userID uuid.UUID, start, end time.Time) (common.Money, error) {
	query := `
		select coalesce(sum(amount), 0)
		from finance_transactions
		where user_id = $1 and type = 'income' and occurred_at >= $2 and occurred_at < $3
	`
	var total common.Money
	err := r.db.QueryRowContext(ctx, query, userID, start, end).Scan(&total)
	return total, err
}

func (r *Repository) MonthlyExpense(ctx context.Context, userID uuid.UUID, start, end time.Time) (common.Money, error) {
	query := `
		select coalesce(sum(amount), 0)
		from finance_transactions
		where user_id = $1 and type = 'expense' and occurred_at >= $2 and occurred_at < $3
	`
	var total common.Money
	err := r.db.QueryRowContext(ctx, query, userID, start, end).Scan(&total)
	return total, err
}

func (r *Repository) TopExpenseCategories(ctx context.Context, userID uuid.UUID, start, end time.Time, limit int) ([]TopCategory, error) {
	query := `
		select c.id::text, coalesce(c.name, 'Без категории') as name, coalesce(sum(t.amount), 0) as amount
		from finance_transactions t
		left join finance_categories c on c.id = t.category_id
		where t.user_id = $1 and t.type = 'expense' and t.occurred_at >= $2 and t.occurred_at < $3
		group by c.id, c.name
		order by amount desc
		limit $4
	`
	rows, err := r.db.QueryContext(ctx, query, userID, start, end, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []TopCategory
	for rows.Next() {
		var item TopCategory
		if err := rows.Scan(&item.CategoryID, &item.Name, &item.Amount); err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, rows.Err()
}
