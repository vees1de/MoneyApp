package sessions

import (
	"context"
	"database/sql"
	"errors"
	"time"

	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/clock"
	platformdb "moneyapp/backend/internal/platform/db"

	"github.com/google/uuid"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
)

type Service struct {
	db         *sql.DB
	repo       *Repository
	jwt        *platformauth.JWTManager
	clock      clock.Clock
	refreshTTL time.Duration
	accessTTL  time.Duration
}

func NewService(database *sql.DB, repo *Repository, jwt *platformauth.JWTManager, clock clock.Clock, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		db:         database,
		repo:       repo,
		jwt:        jwt,
		clock:      clock,
		refreshTTL: refreshTTL,
		accessTTL:  accessTTL,
	}
}

func (s *Service) Issue(ctx context.Context, userID uuid.UUID, meta SessionMeta) (Tokens, Session, error) {
	now := s.clock.Now()
	rawRefresh, err := platformauth.NewOpaqueToken()
	if err != nil {
		return Tokens{}, Session{}, err
	}

	session := Session{
		ID:               uuid.New(),
		UserID:           userID,
		RefreshTokenHash: platformauth.HashToken(rawRefresh),
		UserAgent:        meta.UserAgent,
		IPAddress:        meta.IPAddress,
		ExpiresAt:        now.Add(s.refreshTTL),
		CreatedAt:        now,
	}
	if err := s.repo.Create(ctx, session); err != nil {
		return Tokens{}, Session{}, err
	}

	accessToken, err := s.jwt.SignAccessToken(userID, session.ID, now)
	if err != nil {
		return Tokens{}, Session{}, err
	}

	return Tokens{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, session, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string, meta SessionMeta) (Tokens, Session, error) {
	now := s.clock.Now()
	tokenHash := platformauth.HashToken(refreshToken)
	session, err := s.repo.GetByRefreshTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Tokens{}, Session{}, ErrSessionNotFound
		}
		return Tokens{}, Session{}, err
	}

	if session.RevokedAt != nil || session.ExpiresAt.Before(now) {
		return Tokens{}, Session{}, ErrSessionExpired
	}

	rawRefresh, err := platformauth.NewOpaqueToken()
	if err != nil {
		return Tokens{}, Session{}, err
	}

	session.RefreshTokenHash = platformauth.HashToken(rawRefresh)
	session.UserAgent = meta.UserAgent
	session.IPAddress = meta.IPAddress
	session.ExpiresAt = now.Add(s.refreshTTL)

	err = platformdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		return s.repo.Rotate(ctx, session.ID, session.RefreshTokenHash, session.UserAgent, session.IPAddress, session.ExpiresAt, tx)
	})
	if err != nil {
		return Tokens{}, Session{}, err
	}

	accessToken, err := s.jwt.SignAccessToken(session.UserID, session.ID, now)
	if err != nil {
		return Tokens{}, Session{}, err
	}

	return Tokens{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, session, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := platformauth.HashToken(refreshToken)
	session, err := s.repo.GetByRefreshTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	return s.repo.RevokeByID(ctx, session.ID, s.clock.Now())
}

func (s *Service) RevokeAll(ctx context.Context, userID uuid.UUID) error {
	return s.repo.RevokeAllByUser(ctx, userID, s.clock.Now())
}
