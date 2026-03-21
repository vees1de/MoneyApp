package dashboard

import (
	"context"
	"database/sql"
	"time"

	"moneyapp/backend/internal/core/common"

	"github.com/google/uuid"
)

type CategoryTrend struct {
	Name    string
	Current common.Money
	Average common.Money
}

type Repository struct {
	db *sql.DB
}

func NewRepository(database *sql.DB) *Repository {
	return &Repository{db: database}
}

func (r *Repository) TopExpenseTrend(ctx context.Context, userID uuid.UUID, monthStart, monthEnd time.Time) (CategoryTrend, error) {
	query := `
		with current_month as (
			select coalesce(c.name, 'Без категории') as name, sum(t.amount) as current_amount
			from finance_transactions t
			left join finance_categories c on c.id = t.category_id
			where t.user_id = $1
			  and t.type = 'expense'
			  and t.occurred_at >= $2
			  and t.occurred_at < $3
			group by coalesce(c.name, 'Без категории')
		),
		monthly_history as (
			select date_trunc('month', t.occurred_at) as bucket,
			       coalesce(c.name, 'Без категории') as name,
			       sum(t.amount) as amount
			from finance_transactions t
			left join finance_categories c on c.id = t.category_id
			where t.user_id = $1
			  and t.type = 'expense'
			  and t.occurred_at >= $2 - interval '3 months'
			  and t.occurred_at < $2
			group by 1, 2
		),
		averages as (
			select name, avg(amount) as avg_amount
			from monthly_history
			group by name
		)
		select c.name, c.current_amount, coalesce(a.avg_amount, 0) as avg_amount
		from current_month c
		left join averages a on a.name = c.name
		order by case
			when coalesce(a.avg_amount, 0) = 0 then c.current_amount
			else c.current_amount / a.avg_amount
		end desc
		limit 1
	`

	var trend CategoryTrend
	err := r.db.QueryRowContext(ctx, query, userID, monthStart, monthEnd).Scan(&trend.Name, &trend.Current, &trend.Average)
	return trend, err
}
