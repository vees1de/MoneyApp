package cicd

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

func (r *Repository) Create(ctx context.Context, check SmokeCheck, exec ...db.DBTX) error {
	query := `
		insert into cicd_smoke_checks (
			id, user_id, session_id, request_id, trigger, created_at
		)
		values ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.base(exec...).ExecContext(ctx, query,
		check.ID,
		check.UserID,
		check.SessionID,
		check.RequestID,
		check.Trigger,
		check.CreatedAt,
	)
	return err
}

func (r *Repository) CountByUser(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) (int, error) {
	query := `
		select count(*)
		from cicd_smoke_checks
		where user_id = $1
	`

	var total int
	err := r.base(exec...).QueryRowContext(ctx, query, userID).Scan(&total)
	return total, err
}
