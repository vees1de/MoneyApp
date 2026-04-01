package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"moneyapp/backend/internal/platform/httpx"
)

type LogEntry struct {
	ID          string    `json:"id"`
	ActorUserID *string   `json:"actor_user_id,omitempty"`
	UserID      *string   `json:"user_id,omitempty"`
	EntityType  string    `json:"entity_type"`
	EntityID    *string   `json:"entity_id,omitempty"`
	Action      string    `json:"action"`
	OldValues   any       `json:"old_values,omitempty"`
	NewValues   any       `json:"new_values,omitempty"`
	Meta        any       `json:"meta,omitempty"`
	Source      string    `json:"source"`
	RequestID   *string   `json:"request_id,omitempty"`
	SessionID   *string   `json:"session_id,omitempty"`
	ChangeSet   any       `json:"change_set,omitempty"`
	ActorType   string    `json:"actor_type"`
	ActorID     *string   `json:"actor_id,omitempty"`
	IP          *string   `json:"ip,omitempty"`
	UserAgent   *string   `json:"user_agent,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type Service struct {
	db *sql.DB
}

func NewService(database *sql.DB) *Service {
	return &Service{db: database}
}

func (s *Service) List(ctx context.Context, limit int) ([]LogEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	rows, err := s.db.QueryContext(ctx, `
		select id::text,
		       actor_user_id::text,
		       user_id::text,
		       entity_type,
		       entity_id::text,
		       action,
		       old_values,
		       new_values,
		       meta,
		       source,
		       request_id,
		       session_id::text,
		       change_set,
		       actor_type,
		       actor_id::text,
		       ip::text,
		       user_agent,
		       created_at
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
		var oldValuesBytes []byte
		var newValuesBytes []byte
		var metaBytes []byte
		var changeSetBytes []byte
		if err := rows.Scan(
			&item.ID,
			&item.ActorUserID,
			&item.UserID,
			&item.EntityType,
			&item.EntityID,
			&item.Action,
			&oldValuesBytes,
			&newValuesBytes,
			&metaBytes,
			&item.Source,
			&item.RequestID,
			&item.SessionID,
			&changeSetBytes,
			&item.ActorType,
			&item.ActorID,
			&item.IP,
			&item.UserAgent,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		item.OldValues = decodeJSONValue(oldValuesBytes)
		item.NewValues = decodeJSONValue(newValuesBytes)
		item.Meta = decodeJSONValue(metaBytes)
		item.ChangeSet = decodeJSONValue(changeSetBytes)
		items = append(items, item)
	}
	return items, rows.Err()
}

func decodeJSONValue(payload []byte) any {
	if len(payload) == 0 || string(payload) == "null" {
		return nil
	}

	var value any
	if err := json.Unmarshal(payload, &value); err != nil {
		return map[string]any{"raw": string(payload)}
	}

	return value
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
