package sessions

import (
	"context"
	"database/sql"
	"time"

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

func (r *Repository) Create(ctx context.Context, session Session, exec ...db.DBTX) error {
	query := `
		insert into sessions (
			id, user_id, refresh_token_hash, user_agent, ip_address, expires_at, created_at, revoked_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.base(exec...).ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.RefreshTokenHash,
		session.UserAgent,
		session.IPAddress,
		session.ExpiresAt,
		session.CreatedAt,
		session.RevokedAt,
	)
	return err
}

func (r *Repository) GetByRefreshTokenHash(ctx context.Context, tokenHash string, exec ...db.DBTX) (Session, error) {
	query := `
		select id, user_id, refresh_token_hash, user_agent, ip_address, expires_at, created_at, revoked_at
		from sessions
		where refresh_token_hash = $1
		limit 1
	`

	var session Session
	err := r.base(exec...).QueryRowContext(ctx, query, tokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&session.UserAgent,
		&session.IPAddress,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.RevokedAt,
	)
	return session, err
}

func (r *Repository) RevokeByID(ctx context.Context, sessionID uuid.UUID, revokedAt time.Time, exec ...db.DBTX) error {
	query := `
		update sessions
		set revoked_at = $2
		where id = $1 and revoked_at is null
	`
	_, err := r.base(exec...).ExecContext(ctx, query, sessionID, revokedAt)
	return err
}

func (r *Repository) Rotate(ctx context.Context, sessionID uuid.UUID, refreshTokenHash string, userAgent, ipAddress *string, expiresAt time.Time, exec ...db.DBTX) error {
	query := `
		update sessions
		set refresh_token_hash = $2,
		    user_agent = $3,
		    ip_address = $4,
		    expires_at = $5,
		    revoked_at = null
		where id = $1
	`
	_, err := r.base(exec...).ExecContext(ctx, query, sessionID, refreshTokenHash, userAgent, ipAddress, expiresAt)
	return err
}

func (r *Repository) RevokeAllByUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time, exec ...db.DBTX) error {
	query := `
		update sessions
		set revoked_at = $2
		where user_id = $1 and revoked_at is null
	`
	_, err := r.base(exec...).ExecContext(ctx, query, userID, revokedAt)
	return err
}
