package auth

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

func (r *Repository) FindByProvider(ctx context.Context, provider Provider, providerUserID string, exec ...db.DBTX) (Identity, error) {
	query := `
		select id, user_id, provider, provider_user_id, provider_email, access_meta, created_at
		from auth_identities
		where provider = $1 and provider_user_id = $2
		limit 1
	`

	var identity Identity
	err := r.base(exec...).QueryRowContext(ctx, query, provider, providerUserID).Scan(
		&identity.ID,
		&identity.UserID,
		&identity.Provider,
		&identity.ProviderUserID,
		&identity.ProviderEmail,
		&identity.AccessMeta,
		&identity.CreatedAt,
	)
	return identity, err
}

func (r *Repository) Create(ctx context.Context, identity Identity, exec ...db.DBTX) error {
	query := `
		insert into auth_identities (
			id, user_id, provider, provider_user_id, provider_email, access_meta, created_at
		)
		values ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.base(exec...).ExecContext(ctx, query,
		identity.ID,
		identity.UserID,
		identity.Provider,
		identity.ProviderUserID,
		identity.ProviderEmail,
		identity.AccessMeta,
		identity.CreatedAt,
	)
	return err
}
