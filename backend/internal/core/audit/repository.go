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
			id, actor_user_id, user_id, entity_type, entity_id, action,
			old_values, new_values, meta, source, request_id, session_id,
			change_set, actor_type, actor_id, ip, user_agent, created_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`

	_, err := r.base(exec...).ExecContext(ctx, query,
		event.ID,
		event.ActorUserID,
		event.UserID,
		event.EntityType,
		event.EntityID,
		event.Action,
		event.OldValues,
		event.NewValues,
		event.Meta,
		event.Source,
		event.RequestID,
		event.SessionID,
		event.ChangeSet,
		event.ActorType,
		event.ActorID,
		event.IP,
		event.UserAgent,
		event.CreatedAt,
	)
	return err
}
