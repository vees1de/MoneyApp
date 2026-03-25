package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"moneyapp/backend/internal/config"
	"moneyapp/backend/internal/core/audit"
	"moneyapp/backend/internal/core/sessions"
	"moneyapp/backend/internal/core/users"
	"moneyapp/backend/internal/platform/clock"
	platformdb "moneyapp/backend/internal/platform/db"

	"github.com/google/uuid"
)

type TelegramVerifier interface {
	Verify(context.Context, TelegramVerificationInput) (VerifiedIdentity, error)
}

type YandexVerifier interface {
	Verify(context.Context, YandexVerificationInput) (VerifiedIdentity, error)
}

type Service struct {
	db        *sql.DB
	config    config.AuthConfig
	clock     clock.Clock
	repo      *Repository
	usersRepo *users.Repository
	sessions  *sessions.Service
	audit     *audit.Service
	telegram  TelegramVerifier
	yandex    YandexVerifier
}

func NewService(
	database *sql.DB,
	cfg config.AuthConfig,
	clock clock.Clock,
	repo *Repository,
	usersRepo *users.Repository,
	sessionsService *sessions.Service,
	auditService *audit.Service,
	telegramVerifier TelegramVerifier,
	yandexVerifier YandexVerifier,
) *Service {
	return &Service{
		db:        database,
		config:    cfg,
		clock:     clock,
		repo:      repo,
		usersRepo: usersRepo,
		sessions:  sessionsService,
		audit:     auditService,
		telegram:  telegramVerifier,
		yandex:    yandexVerifier,
	}
}

func (s *Service) LoginWithTelegram(ctx context.Context, request TelegramLoginRequest, meta sessions.SessionMeta) (AuthResponse, error) {
	verified, err := s.telegram.Verify(ctx, TelegramVerificationInput{
		IDToken: request.IDToken,
		Nonce:   request.Nonce,
	})
	if err != nil {
		return AuthResponse{}, err
	}

	return s.loginWithVerifiedIdentity(ctx, verified, meta)
}

func (s *Service) LoginWithYandex(ctx context.Context, request YandexLoginRequest, meta sessions.SessionMeta) (AuthResponse, error) {
	verified, err := s.yandex.Verify(ctx, YandexVerificationInput{
		Code:           request.Code,
		IDToken:        request.IDToken,
		ProviderUserID: request.ProviderUserID,
		Email:          request.Email,
		DisplayName:    request.DisplayName,
		AvatarURL:      request.AvatarURL,
	})
	if err != nil {
		return AuthResponse{}, err
	}

	return s.loginWithVerifiedIdentity(ctx, verified, meta)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string, meta sessions.SessionMeta) (AuthResponse, error) {
	tokens, refreshedSession, err := s.sessions.Refresh(ctx, refreshToken, meta)
	if err != nil {
		return AuthResponse{}, err
	}

	user, err := s.usersRepo.GetByID(ctx, refreshedSession.UserID)
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{
		User:   user,
		Tokens: tokens,
	}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	return s.sessions.Logout(ctx, refreshToken)
}

func (s *Service) Me(ctx context.Context, userID uuid.UUID) (users.User, error) {
	return s.usersRepo.GetByID(ctx, userID)
}

func (s *Service) loginWithVerifiedIdentity(ctx context.Context, verified VerifiedIdentity, meta sessions.SessionMeta) (AuthResponse, error) {
	var user users.User
	var created bool

	err := platformdb.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		identity, err := s.repo.FindByProvider(ctx, verified.Provider, verified.ProviderUserID, tx)
		switch {
		case err == nil:
			user, err = s.usersRepo.GetByID(ctx, identity.UserID, tx)
			return err
		case !errors.Is(err, sql.ErrNoRows):
			return err
		}

		if verified.Email != nil {
			existing, err := s.usersRepo.GetByEmail(ctx, *verified.Email, tx)
			switch {
			case err == nil:
				user = existing
			case !errors.Is(err, sql.ErrNoRows):
				return err
			default:
				user = s.newUser(verified)
				created = true
				if err := s.usersRepo.Create(ctx, user, tx); err != nil {
					return err
				}
			}
		} else {
			user = s.newUser(verified)
			created = true
			if err := s.usersRepo.Create(ctx, user, tx); err != nil {
				return err
			}
		}

		payload, err := json.Marshal(verified.AccessMeta)
		if err != nil {
			return err
		}

		return s.repo.Create(ctx, Identity{
			ID:             uuid.New(),
			UserID:         user.ID,
			Provider:       verified.Provider,
			ProviderUserID: verified.ProviderUserID,
			ProviderEmail:  verified.Email,
			AccessMeta:     payload,
			CreatedAt:      s.clock.Now(),
		}, tx)
	})
	if err != nil {
		return AuthResponse{}, err
	}

	tokens, _, err := s.sessions.Issue(ctx, user.ID, meta)
	if err != nil {
		return AuthResponse{}, err
	}

	action := "auth.login"
	if created {
		action = "auth.signup"
	}
	if err := s.audit.Record(ctx, user.ID, action, "user", &user.ID, map[string]any{
		"provider": string(verified.Provider),
	}); err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{
		User:   user,
		Tokens: tokens,
		Meta: map[string]any{
			"created": created,
		},
	}, nil
}

func (s *Service) newUser(verified VerifiedIdentity) users.User {
	displayName := verified.DisplayName
	if displayName == nil && verified.Email != nil {
		localPart := strings.Split(*verified.Email, "@")[0]
		displayName = &localPart
	}

	now := s.clock.Now()
	return users.User{
		ID:                  uuid.New(),
		Email:               verified.Email,
		DisplayName:         displayName,
		AvatarURL:           verified.AvatarURL,
		Timezone:            s.config.DefaultTimezone,
		BaseCurrency:        s.config.DefaultBaseCurrency,
		OnboardingCompleted: false,
		WeeklyReviewWeekday: 1,
		WeeklyReviewHour:    s.config.DefaultWeeklyReviewHour,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

var ErrProviderIdentityInvalid = fmt.Errorf("provider identity invalid")
