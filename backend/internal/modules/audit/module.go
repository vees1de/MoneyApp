package audit

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"moneyapp/backend/internal/platform/httpx"
)

type LogEntry struct {
	ID          string         `json:"id"`
	ActorUserID *string        `json:"actor_user_id,omitempty"`
	UserID      *string        `json:"user_id,omitempty"`
	EntityType  string         `json:"entity_type"`
	EntityID    *string        `json:"entity_id,omitempty"`
	Action      string         `json:"action"`
	Meta        map[string]any `json:"meta,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

type Service struct {
	db *sql.DB
}

func NewService(database *sql.DB) *Service {
	return &Service{db: database}
}

func (s *Service) List(ctx context.Context, limit int) ([]LogEntry, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
		select id::text, actor_user_id::text, user_id::text, entity_type, entity_id::text, action, meta, created_at
		from audit_logs
		order by created_at desc
		limit $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []LogEntry
	for rows.Next() {
		var item LogEntry
		var metaBytes []byte
		if err := rows.Scan(&item.ID, &item.ActorUserID, &item.UserID, &item.EntityType, &item.EntityID, &item.Action, &metaBytes, &item.CreatedAt); err != nil {
			return nil, err
		}
		if len(metaBytes) > 0 {
			item.Meta = map[string]any{"raw": string(metaBytes)}
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.service.List(r.Context(), limit)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}
