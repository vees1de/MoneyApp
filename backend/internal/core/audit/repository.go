package audit

import (
	"context"
	"database/sql"

	"moneyapp/backend/internal/platform/db"
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

func (r *Repository) Create(ctx context.Context, event Event, exec ...db.DBTX) error {
	query := `
		insert into audit_logs (
			id, user_id, action, entity_type, entity_id, meta, source,
			request_id, session_id, change_set, actor_type, actor_id, created_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.base(exec...).ExecContext(ctx, query,
		event.ID,
		event.UserID,
		event.Action,
		event.EntityType,
		event.EntityID,
		event.Meta,
		event.Source,
		event.RequestID,
		event.SessionID,
		event.ChangeSet,
		event.ActorType,
		event.ActorID,
		event.CreatedAt,
	)
	return err
}
