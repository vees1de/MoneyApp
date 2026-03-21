package review

import (
	"context"
	"database/sql"

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

func (r *Repository) FindByPeriod(ctx context.Context, userID uuid.UUID, periodStart, periodEnd string, exec ...db.DBTX) (WeeklyReview, error) {
	query := `
		select id, user_id, account_id, period_start, period_end, expected_balance, actual_balance, delta, status, resolution_note, created_at, completed_at
		from weekly_reviews
		where user_id = $1 and period_start = $2 and period_end = $3
		limit 1
	`
	var review WeeklyReview
	err := r.base(exec...).QueryRowContext(ctx, query, userID, periodStart, periodEnd).Scan(
		&review.ID,
		&review.UserID,
		&review.AccountID,
		&review.PeriodStart,
		&review.PeriodEnd,
		&review.ExpectedBalance,
		&review.ActualBalance,
		&review.Delta,
		&review.Status,
		&review.ResolutionNote,
		&review.CreatedAt,
		&review.CompletedAt,
	)
	return review, err
}

func (r *Repository) GetByID(ctx context.Context, userID, reviewID uuid.UUID, exec ...db.DBTX) (WeeklyReview, error) {
	query := `
		select id, user_id, account_id, period_start, period_end, expected_balance, actual_balance, delta, status, resolution_note, created_at, completed_at
		from weekly_reviews
		where id = $1 and user_id = $2
	`
	var review WeeklyReview
	err := r.base(exec...).QueryRowContext(ctx, query, reviewID, userID).Scan(
		&review.ID,
		&review.UserID,
		&review.AccountID,
		&review.PeriodStart,
		&review.PeriodEnd,
		&review.ExpectedBalance,
		&review.ActualBalance,
		&review.Delta,
		&review.Status,
		&review.ResolutionNote,
		&review.CreatedAt,
		&review.CompletedAt,
	)
	return review, err
}

func (r *Repository) Create(ctx context.Context, review WeeklyReview, exec ...db.DBTX) error {
	query := `
		insert into weekly_reviews (
			id, user_id, account_id, period_start, period_end, expected_balance, actual_balance, delta, status, resolution_note, created_at, completed_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.base(exec...).ExecContext(ctx, query,
		review.ID,
		review.UserID,
		review.AccountID,
		review.PeriodStart,
		review.PeriodEnd,
		review.ExpectedBalance,
		review.ActualBalance,
		review.Delta,
		review.Status,
		review.ResolutionNote,
		review.CreatedAt,
		review.CompletedAt,
	)
	return err
}

func (r *Repository) Update(ctx context.Context, review WeeklyReview, exec ...db.DBTX) error {
	query := `
		update weekly_reviews
		set actual_balance = $3,
		    delta = $4,
		    status = $5,
		    resolution_note = $6,
		    completed_at = $7
		where id = $1 and user_id = $2
	`
	_, err := r.base(exec...).ExecContext(ctx, query,
		review.ID,
		review.UserID,
		review.ActualBalance,
		review.Delta,
		review.Status,
		review.ResolutionNote,
		review.CompletedAt,
	)
	return err
}
