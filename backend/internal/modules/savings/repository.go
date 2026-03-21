package savings

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

func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) ([]Goal, error) {
	query := `
		select id, user_id, title, target_amount, current_amount, currency, target_date, priority, status, created_at, updated_at
		from savings_goals
		where user_id = $1
		order by created_at asc
	`
	rows, err := r.base(exec...).QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Goal
	for rows.Next() {
		item, err := scanGoal(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) ListActiveByUser(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) ([]Goal, error) {
	query := `
		select id, user_id, title, target_amount, current_amount, currency, target_date, priority, status, created_at, updated_at
		from savings_goals
		where user_id = $1 and status = 'active'
		order by priority desc, created_at asc
	`
	rows, err := r.base(exec...).QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Goal
	for rows.Next() {
		item, err := scanGoal(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, userID, goalID uuid.UUID, exec ...db.DBTX) (Goal, error) {
	query := `
		select id, user_id, title, target_amount, current_amount, currency, target_date, priority, status, created_at, updated_at
		from savings_goals
		where id = $1 and user_id = $2
	`
	var goal Goal
	err := r.base(exec...).QueryRowContext(ctx, query, goalID, userID).Scan(
		&goal.ID,
		&goal.UserID,
		&goal.Title,
		&goal.TargetAmount,
		&goal.CurrentAmount,
		&goal.Currency,
		&goal.TargetDate,
		&goal.Priority,
		&goal.Status,
		&goal.CreatedAt,
		&goal.UpdatedAt,
	)
	return goal, err
}

func (r *Repository) Create(ctx context.Context, goal Goal, exec ...db.DBTX) error {
	query := `
		insert into savings_goals (
			id, user_id, title, target_amount, current_amount, currency, target_date, priority, status, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.base(exec...).ExecContext(ctx, query,
		goal.ID,
		goal.UserID,
		goal.Title,
		goal.TargetAmount,
		goal.CurrentAmount,
		goal.Currency,
		goal.TargetDate,
		goal.Priority,
		goal.Status,
		goal.CreatedAt,
		goal.UpdatedAt,
	)
	return err
}

func (r *Repository) Update(ctx context.Context, goal Goal, exec ...db.DBTX) error {
	query := `
		update savings_goals
		set title = $3,
		    target_amount = $4,
		    current_amount = $5,
		    target_date = $6,
		    priority = $7,
		    status = $8,
		    updated_at = $9
		where id = $1 and user_id = $2
	`
	_, err := r.base(exec...).ExecContext(ctx, query,
		goal.ID,
		goal.UserID,
		goal.Title,
		goal.TargetAmount,
		goal.CurrentAmount,
		goal.TargetDate,
		goal.Priority,
		goal.Status,
		goal.UpdatedAt,
	)
	return err
}

type goalScanner interface {
	Scan(dest ...any) error
}

func scanGoal(scanner goalScanner) (Goal, error) {
	var goal Goal
	err := scanner.Scan(
		&goal.ID,
		&goal.UserID,
		&goal.Title,
		&goal.TargetAmount,
		&goal.CurrentAmount,
		&goal.Currency,
		&goal.TargetDate,
		&goal.Priority,
		&goal.Status,
		&goal.CreatedAt,
		&goal.UpdatedAt,
	)
	return goal, err
}
