package outlook

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/clock"
	"moneyapp/backend/internal/platform/db"
	"moneyapp/backend/internal/platform/httpx"
	"moneyapp/backend/internal/platform/worker"

	"github.com/google/uuid"
)

type Account struct {
	ID                uuid.UUID  `json:"id"`
	UserID            uuid.UUID  `json:"user_id"`
	ExternalAccountID string     `json:"external_account_id"`
	Email             string     `json:"email"`
	AccessToken       string     `json:"-"`
	RefreshToken      string     `json:"-"`
	TokenExpiresAt    time.Time  `json:"token_expires_at"`
	Scope             *string    `json:"scope,omitempty"`
	Status            string     `json:"status"`
	LastSyncAt        *time.Time `json:"last_sync_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type IntegrationStatus struct {
	Connected bool     `json:"connected"`
	Account   *Account `json:"account,omitempty"`
}

type ConnectResponse struct {
	AuthURL string `json:"auth_url"`
	State   string `json:"state"`
}

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

func (r *Repository) GetByUserID(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) (Account, error) {
	var item Account
	err := r.base(exec...).QueryRowContext(ctx, `
		select id, user_id, external_account_id, email, access_token_encrypted, refresh_token_encrypted,
		       token_expires_at, scope, status, last_sync_at, created_at, updated_at
		from outlook_accounts
		where user_id = $1
	`, userID).Scan(&item.ID, &item.UserID, &item.ExternalAccountID, &item.Email, &item.AccessToken, &item.RefreshToken,
		&item.TokenExpiresAt, &item.Scope, &item.Status, &item.LastSyncAt, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repository) UpsertAccount(ctx context.Context, item Account, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into outlook_accounts (
			id, user_id, external_account_id, email, access_token_encrypted, refresh_token_encrypted,
			token_expires_at, scope, status, last_sync_at, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		on conflict (user_id) do update
		set external_account_id = excluded.external_account_id,
		    email = excluded.email,
		    access_token_encrypted = excluded.access_token_encrypted,
		    refresh_token_encrypted = excluded.refresh_token_encrypted,
		    token_expires_at = excluded.token_expires_at,
		    scope = excluded.scope,
		    status = excluded.status,
		    last_sync_at = excluded.last_sync_at,
		    updated_at = excluded.updated_at
	`, item.ID, item.UserID, item.ExternalAccountID, item.Email, item.AccessToken, item.RefreshToken,
		item.TokenExpiresAt, item.Scope, item.Status, item.LastSyncAt, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repository) Disconnect(ctx context.Context, userID uuid.UUID, updatedAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update outlook_accounts
		set status = 'revoked', updated_at = $2
		where user_id = $1
	`, userID, updatedAt)
	return err
}

func (r *Repository) CreateIntegrationJob(ctx context.Context, userID uuid.UUID, payload string, createdAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into integration_jobs (
			id, integration_type, entity_type, entity_id, job_type, status, attempt, max_attempts,
			next_retry_at, last_error, payload, created_at, updated_at
		)
		values ($1, 'outlook_calendar_sync', 'user', $2, 'pull_changes', 'pending', 0, 5, null, null, $3::jsonb, $4, $4)
	`, uuid.New(), userID, payload, createdAt)
	return err
}

type Service struct {
	db    *sql.DB
	repo  *Repository
	queue *worker.Queue
	clock clock.Clock
}

func NewService(database *sql.DB, repo *Repository, queue *worker.Queue, appClock clock.Clock) *Service {
	return &Service{
		db:    database,
		repo:  repo,
		queue: queue,
		clock: appClock,
	}
}

func (s *Service) Connect(principal platformauth.Principal) ConnectResponse {
	state := principal.UserID.String()
	return ConnectResponse{
		AuthURL: "/api/v1/integrations/outlook/callback?state=" + state,
		State:   state,
	}
}

func (s *Service) Callback(ctx context.Context, state, email, externalAccountID, code string) (Account, error) {
	userID, err := uuid.Parse(strings.TrimSpace(state))
	if err != nil {
		return Account{}, httpx.BadRequest("invalid_state", "invalid outlook state")
	}
	if strings.TrimSpace(email) == "" {
		email = "outlook+" + userID.String() + "@example.local"
	}
	if strings.TrimSpace(externalAccountID) == "" {
		externalAccountID = "outlook-" + userID.String()
	}
	if strings.TrimSpace(code) == "" {
		code = uuid.NewString()
	}

	now := s.clock.Now()
	account := Account{
		ID:                uuid.New(),
		UserID:            userID,
		ExternalAccountID: externalAccountID,
		Email:             email,
		AccessToken:       code,
		RefreshToken:      "refresh-" + code,
		TokenExpiresAt:    now.Add(1 * time.Hour),
		Status:            "active",
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := s.repo.UpsertAccount(ctx, account); err != nil {
		return Account{}, err
	}
	if err := s.queue.Enqueue(ctx, s.db, "integrations", "outlook_sync", map[string]any{
		"user_id": userID,
	}, nil, now); err != nil {
		return Account{}, err
	}
	return account, nil
}

func (s *Service) Disconnect(ctx context.Context, principal platformauth.Principal) error {
	return s.repo.Disconnect(ctx, principal.UserID, s.clock.Now())
}

func (s *Service) Status(ctx context.Context, principal platformauth.Principal) (IntegrationStatus, error) {
	account, err := s.repo.GetByUserID(ctx, principal.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return IntegrationStatus{Connected: false}, nil
		}
		return IntegrationStatus{}, err
	}
	return IntegrationStatus{Connected: account.Status == "active", Account: &account}, nil
}

func (s *Service) Sync(ctx context.Context, principal platformauth.Principal) error {
	now := s.clock.Now()
	if err := s.repo.CreateIntegrationJob(ctx, principal.UserID, `{"source":"manual_sync"}`, now); err != nil {
		return err
	}
	return s.queue.Enqueue(ctx, s.db, "integrations", "outlook_sync", map[string]any{
		"user_id": principal.UserID,
	}, nil, now)
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func outlookPrincipal(r *http.Request) (platformauth.Principal, error) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		return platformauth.Principal{}, httpx.Unauthorized("unauthorized", "authorization required")
	}
	return principal, nil
}

func (h *Handler) Connect(w http.ResponseWriter, r *http.Request) {
	principal, err := outlookPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, h.service.Connect(principal))
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	account, err := h.service.Callback(r.Context(), r.URL.Query().Get("state"), r.URL.Query().Get("email"), r.URL.Query().Get("account_id"), r.URL.Query().Get("code"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, account)
}

func (h *Handler) Disconnect(w http.ResponseWriter, r *http.Request) {
	principal, err := outlookPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.service.Disconnect(r.Context(), principal); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	principal, err := outlookPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	status, err := h.service.Status(r.Context(), principal)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, status)
}

func (h *Handler) Sync(w http.ResponseWriter, r *http.Request) {
	principal, err := outlookPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.service.Sync(r.Context(), principal); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteNoContent(w)
}
