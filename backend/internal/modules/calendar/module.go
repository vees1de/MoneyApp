package calendar

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/google/uuid"
)

const (
	defaultUpcomingLimit = 5
	maxUpcomingLimit     = 20
)

type UpcomingCalendarEvent struct {
	ID              uuid.UUID `json:"id"`
	SourceType      string    `json:"source_type"`
	SourceID        uuid.UUID `json:"source_id"`
	Provider        string    `json:"provider"`
	ExternalEventID *string   `json:"external_event_id,omitempty"`
	Title           string    `json:"title"`
	StartAt         time.Time `json:"start_at"`
	EndAt           time.Time `json:"end_at"`
	Timezone        *string   `json:"timezone,omitempty"`
	Status          string    `json:"status"`
	MeetingURL      *string   `json:"meeting_url,omitempty"`
	Location        *string   `json:"location,omitempty"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(database *sql.DB) *Repository {
	return &Repository{db: database}
}

func (r *Repository) ListUpcoming(ctx context.Context, userID uuid.UUID, now time.Time, limit int) ([]UpcomingCalendarEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
		select id, source_type, source_id, provider, external_event_id, title, start_at, end_at, timezone, status, meeting_url, location
		from calendar_events
		where user_id = $1
		  and status in ('scheduled', 'updated')
		  and end_at >= $2
		order by start_at asc
		limit $3
	`, userID, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []UpcomingCalendarEvent
	for rows.Next() {
		var item UpcomingCalendarEvent
		if err := rows.Scan(&item.ID, &item.SourceType, &item.SourceID, &item.Provider, &item.ExternalEventID, &item.Title, &item.StartAt, &item.EndAt, &item.Timezone, &item.Status, &item.MeetingURL, &item.Location); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListUpcoming(ctx context.Context, principal platformauth.Principal, limit int) ([]UpcomingCalendarEvent, error) {
	if limit <= 0 {
		limit = defaultUpcomingLimit
	}
	if limit > maxUpcomingLimit {
		limit = maxUpcomingLimit
	}
	return s.repo.ListUpcoming(ctx, principal.UserID, time.Now().UTC(), limit)
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Upcoming(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}
	limit := defaultUpcomingLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	items, err := h.service.ListUpcoming(r.Context(), principal, limit)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}
