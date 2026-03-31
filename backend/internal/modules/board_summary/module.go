package board_summary

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/google/uuid"
)

type Summary struct {
	BoardsTotal    int `json:"boards_total"`
	TasksTotal     int `json:"tasks_total"`
	ActiveTotal    int `json:"active_total"`
	CompletedTotal int `json:"completed_total"`
	OverdueTotal   int `json:"overdue_total"`
}

type BoardItem struct {
	BoardID        string `json:"board_id"`
	Title          string `json:"title"`
	TasksTotal     int    `json:"tasks_total"`
	ActiveTotal    int    `json:"active_total"`
	CompletedTotal int    `json:"completed_total"`
	OverdueTotal   int    `json:"overdue_total"`
}

type OverdueTask struct {
	TaskID     string     `json:"task_id"`
	BoardID    *string    `json:"board_id,omitempty"`
	BoardTitle *string    `json:"board_title,omitempty"`
	Title      string     `json:"title"`
	DeadlineAt *time.Time `json:"deadline_at,omitempty"`
	Completed  bool       `json:"completed"`
	Archived   bool       `json:"archived"`
}

type BoardSummary struct {
	Source       string        `json:"source"`
	Status       string        `json:"status"`
	Summary      Summary       `json:"summary"`
	Boards       []BoardItem   `json:"boards"`
	OverdueTasks []OverdueTask `json:"overdue_tasks"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(database *sql.DB) *Repository {
	return &Repository{db: database}
}

func (r *Repository) resolveConnection(ctx context.Context, userID uuid.UUID, connectionID *uuid.UUID) (*uuid.UUID, error) {
	query := `select id from integration_yougile_connections where created_by = $1 and status <> 'revoked'`
	args := []any{userID}
	if connectionID != nil {
		query += ` and id = $2`
		args = append(args, *connectionID)
	}
	query += ` order by updated_at desc limit 1`

	var id uuid.UUID
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}

func (r *Repository) listBoards(ctx context.Context, connectionID uuid.UUID, boardID *string) ([]BoardItem, error) {
	args := []any{connectionID, time.Now().UTC()}
	query := `
		select b.yougile_board_id,
		       b.title,
		       count(t.id)::int as tasks_total,
		       count(*) filter (where coalesce(t.completed, false) = false and coalesce(t.archived, false) = false)::int as active_total,
		       count(*) filter (where coalesce(t.completed, false) = true and coalesce(t.archived, false) = false)::int as completed_total,
		       count(*) filter (where coalesce(t.completed, false) = false and coalesce(t.archived, false) = false and t.deadline_at is not null and t.deadline_at < $2)::int as overdue_total
		from yougile_boards b
		left join yougile_tasks t on t.connection_id = b.connection_id and t.yougile_board_id = b.yougile_board_id
		where b.connection_id = $1 and b.deleted = false
	`
	if boardID != nil && *boardID != "" {
		args = append(args, *boardID)
		query += ` and b.yougile_board_id = $3`
	}
	query += ` group by b.yougile_board_id, b.title order by b.title asc`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []BoardItem
	for rows.Next() {
		var item BoardItem
		if err := rows.Scan(&item.BoardID, &item.Title, &item.TasksTotal, &item.ActiveTotal, &item.CompletedTotal, &item.OverdueTotal); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) listOverdueTasks(ctx context.Context, connectionID uuid.UUID, boardID *string) ([]OverdueTask, error) {
	args := []any{connectionID, time.Now().UTC()}
	query := `
		select t.yougile_task_id, t.yougile_board_id, b.title, t.title, t.deadline_at, t.completed, t.archived
		from yougile_tasks t
		left join yougile_boards b on b.connection_id = t.connection_id and b.yougile_board_id = t.yougile_board_id
		where t.connection_id = $1
		  and coalesce(t.completed, false) = false
		  and coalesce(t.archived, false) = false
		  and t.deadline_at is not null
		  and t.deadline_at < $2
	`
	if boardID != nil && *boardID != "" {
		args = append(args, *boardID)
		query += ` and t.yougile_board_id = $3`
	}
	query += ` order by t.deadline_at asc limit 5`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []OverdueTask
	for rows.Next() {
		var item OverdueTask
		if err := rows.Scan(&item.TaskID, &item.BoardID, &item.BoardTitle, &item.Title, &item.DeadlineAt, &item.Completed, &item.Archived); err != nil {
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

func (s *Service) Summary(ctx context.Context, principal platformauth.Principal, connectionID *uuid.UUID, boardID *string) (BoardSummary, error) {
	resolvedID, err := s.repo.resolveConnection(ctx, principal.UserID, connectionID)
	if err != nil {
		return BoardSummary{}, err
	}
	if resolvedID == nil {
		return BoardSummary{Source: "yougile", Status: "unavailable", Boards: []BoardItem{}, OverdueTasks: []OverdueTask{}}, nil
	}

	boards, err := s.repo.listBoards(ctx, *resolvedID, boardID)
	if err != nil {
		return BoardSummary{}, err
	}
	if len(boards) == 0 {
		return BoardSummary{Source: "yougile", Status: "unavailable", Boards: []BoardItem{}, OverdueTasks: []OverdueTask{}}, nil
	}

	overdueTasks, err := s.repo.listOverdueTasks(ctx, *resolvedID, boardID)
	if err != nil {
		return BoardSummary{}, err
	}

	payload := BoardSummary{
		Source:       "yougile",
		Status:       "ready",
		Boards:       boards,
		OverdueTasks: overdueTasks,
	}
	for _, item := range boards {
		payload.Summary.BoardsTotal++
		payload.Summary.TasksTotal += item.TasksTotal
		payload.Summary.ActiveTotal += item.ActiveTotal
		payload.Summary.CompletedTotal += item.CompletedTotal
		payload.Summary.OverdueTotal += item.OverdueTotal
	}
	return payload, nil
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}
	var connectionID *uuid.UUID
	if raw := r.URL.Query().Get("connection_id"); raw != "" {
		value, err := uuid.Parse(raw)
		if err != nil {
			httpx.WriteError(w, httpx.BadRequest("invalid_connection_id", "connection_id must be a valid UUID"))
			return
		}
		connectionID = &value
	}
	var boardID *string
	if raw := r.URL.Query().Get("board_id"); raw != "" {
		value := raw
		boardID = &value
	}
	payload, err := h.service.Summary(r.Context(), principal, connectionID, boardID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, payload)
}
