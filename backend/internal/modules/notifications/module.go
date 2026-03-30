package notifications

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/clock"
	"moneyapp/backend/internal/platform/db"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Notification struct {
	ID                uuid.UUID  `json:"id"`
	UserID            uuid.UUID  `json:"user_id"`
	Channel           string     `json:"channel"`
	Type              string     `json:"type"`
	Title             string     `json:"title"`
	Body              string     `json:"body"`
	Status            string     `json:"status"`
	RelatedEntityType *string    `json:"related_entity_type,omitempty"`
	RelatedEntityID   *uuid.UUID `json:"related_entity_id,omitempty"`
	ScheduledAt       *time.Time `json:"scheduled_at,omitempty"`
	SentAt            *time.Time `json:"sent_at,omitempty"`
	ReadAt            *time.Time `json:"read_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
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

func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) ([]Notification, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select id, user_id, channel, type, title, body, status,
		       related_entity_type, related_entity_id, scheduled_at, sent_at, read_at, created_at
		from notifications
		where user_id = $1
		order by created_at desc
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Notification
	for rows.Next() {
		var item Notification
		if err := rows.Scan(&item.ID, &item.UserID, &item.Channel, &item.Type, &item.Title, &item.Body, &item.Status,
			&item.RelatedEntityType, &item.RelatedEntityID, &item.ScheduledAt, &item.SentAt, &item.ReadAt, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) MarkRead(ctx context.Context, id, userID uuid.UUID, readAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update notifications
		set status = 'read', read_at = $3
		where id = $1 and user_id = $2
	`, id, userID, readAt)
	return err
}

func (r *Repository) MarkAllRead(ctx context.Context, userID uuid.UUID, readAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update notifications
		set status = 'read', read_at = $2
		where user_id = $1 and status <> 'read'
	`, userID, readAt)
	return err
}

type Service struct {
	repo  *Repository
	clock clock.Clock
}

func NewService(repo *Repository, appClock clock.Clock) *Service {
	return &Service{repo: repo, clock: appClock}
}

func (s *Service) ListMine(ctx context.Context, principal platformauth.Principal) ([]Notification, error) {
	return s.repo.ListByUser(ctx, principal.UserID)
}

func (s *Service) MarkRead(ctx context.Context, principal platformauth.Principal, notificationID uuid.UUID) error {
	return s.repo.MarkRead(ctx, notificationID, principal.UserID, s.clock.Now())
}

func (s *Service) MarkAllRead(ctx context.Context, principal platformauth.Principal) error {
	return s.repo.MarkAllRead(ctx, principal.UserID, s.clock.Now())
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func notificationsPrincipal(r *http.Request) (platformauth.Principal, error) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		return platformauth.Principal{}, httpx.Unauthorized("unauthorized", "authorization required")
	}
	return principal, nil
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	principal, err := notificationsPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	items, err := h.service.ListMine(r.Context(), principal)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	principal, err := notificationsPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_notification_id", "invalid notification id"))
		return
	}
	if err := h.service.MarkRead(r.Context(), principal, id); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	principal, err := notificationsPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.service.MarkAllRead(r.Context(), principal); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteNoContent(w)
}
