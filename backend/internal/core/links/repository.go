package links

import (
	"context"
	"database/sql"
	"encoding/json"

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

func (r *Repository) Create(ctx context.Context, link EntityLink, exec ...db.DBTX) error {
	query := `
		insert into entity_links (
			id, user_id, source_type, source_id, target_type, target_id, relation, meta, created_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.base(exec...).ExecContext(ctx, query,
		link.ID,
		link.UserID,
		link.SourceType,
		link.SourceID,
		link.TargetType,
		link.TargetID,
		link.Relation,
		link.Meta,
		link.CreatedAt,
	)
	return err
}

func (r *Repository) ListByEntity(ctx context.Context, userID uuid.UUID, entityType string, entityID uuid.UUID, exec ...db.DBTX) ([]EntityLink, error) {
	query := `
		select id, user_id, source_type, source_id, target_type, target_id, relation, meta, created_at
		from entity_links
		where user_id = $1
		  and (
		        (source_type = $2 and source_id = $3)
		     or (target_type = $2 and target_id = $3)
		  )
		order by created_at desc
	`

	rows, err := r.base(exec...).QueryContext(ctx, query, userID, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []EntityLink
	for rows.Next() {
		var link EntityLink
		var meta []byte
		if err := rows.Scan(
			&link.ID,
			&link.UserID,
			&link.SourceType,
			&link.SourceID,
			&link.TargetType,
			&link.TargetID,
			&link.Relation,
			&meta,
			&link.CreatedAt,
		); err != nil {
			return nil, err
		}
		link.Meta = json.RawMessage(meta)
		result = append(result, link)
	}

	return result, rows.Err()
}
