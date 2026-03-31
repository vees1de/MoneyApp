package yougile

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/clock"
	"moneyapp/backend/internal/platform/db"
	"moneyapp/backend/internal/platform/httpx"
	"moneyapp/backend/internal/platform/worker"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Connection struct {
	ID                uuid.UUID  `json:"id"`
	CompanyID         *uuid.UUID `json:"company_id,omitempty"`
	Title             *string    `json:"title,omitempty"`
	APIBaseURL        string     `json:"api_base_url"`
	YougileCompanyID  string     `json:"yougile_company_id"`
	APIKeyLast4       *string    `json:"api_key_last4,omitempty"`
	Status            string     `json:"status"`
	CreatedBy         uuid.UUID  `json:"created_by"`
	LastSyncAt        *time.Time `json:"last_sync_at,omitempty"`
	LastSuccessSyncAt *time.Time `json:"last_success_sync_at,omitempty"`
	LastError         *string    `json:"last_error,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type storedConnection struct {
	Connection
	APIKeyEncrypted string
}

type ImportedUser struct {
	ID             uuid.UUID  `json:"id"`
	ConnectionID   uuid.UUID  `json:"connection_id"`
	YougileUserID  string     `json:"yougile_user_id"`
	Email          *string    `json:"email,omitempty"`
	RealName       *string    `json:"real_name,omitempty"`
	IsAdmin        bool       `json:"is_admin"`
	Status         *string    `json:"status,omitempty"`
	LastActivityAt *time.Time `json:"last_activity_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type Project struct {
	ID               uuid.UUID `json:"id"`
	ConnectionID     uuid.UUID `json:"connection_id"`
	YougileProjectID string    `json:"yougile_project_id"`
	Title            string    `json:"title"`
	Deleted          bool      `json:"deleted"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Board struct {
	ID               uuid.UUID `json:"id"`
	ConnectionID     uuid.UUID `json:"connection_id"`
	YougileBoardID   string    `json:"yougile_board_id"`
	YougileProjectID string    `json:"yougile_project_id"`
	Title            string    `json:"title"`
	Deleted          bool      `json:"deleted"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Column struct {
	ID              uuid.UUID `json:"id"`
	ConnectionID    uuid.UUID `json:"connection_id"`
	YougileColumnID string    `json:"yougile_column_id"`
	YougileBoardID  string    `json:"yougile_board_id"`
	Title           string    `json:"title"`
	Color           *int      `json:"color,omitempty"`
	Deleted         bool      `json:"deleted"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type MappingView struct {
	ID              uuid.UUID `json:"id"`
	EmployeeUserID  uuid.UUID `json:"employee_user_id"`
	EmployeeName    string    `json:"employee_name"`
	EmployeeEmail   string    `json:"employee_email"`
	YougileUserID   string    `json:"yougile_user_id"`
	YougileRealName *string   `json:"yougile_real_name,omitempty"`
	YougileEmail    *string   `json:"yougile_email,omitempty"`
	MatchSource     string    `json:"match_source"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type SyncJob struct {
	ID           uuid.UUID       `json:"id"`
	ConnectionID uuid.UUID       `json:"connection_id"`
	JobType      string          `json:"job_type"`
	Status       string          `json:"status"`
	Cursor       json.RawMessage `json:"cursor,omitempty"`
	Progress     json.RawMessage `json:"progress"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	FinishedAt   *time.Time      `json:"finished_at,omitempty"`
	Attempt      int             `json:"attempt"`
	NextRetryAt  *time.Time      `json:"next_retry_at,omitempty"`
	ErrorText    *string         `json:"error_text,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type CreateConnectionRequest struct {
	Title            *string `json:"title,omitempty"`
	APIBaseURL       string  `json:"apiBaseUrl" validate:"omitempty,url"`
	AuthMode         string  `json:"authMode" validate:"required,oneof=api_key"`
	YougileCompanyID string  `json:"yougileCompanyId" validate:"required"`
	APIKey           string  `json:"apiKey" validate:"required"`
}

type TestKeyRequest struct {
	APIBaseURL string `json:"apiBaseUrl" validate:"omitempty,url"`
	APIKey     string `json:"apiKey" validate:"required"`
}

type CreateKeyRequest struct {
	Login     string `json:"login" validate:"required"`
	Password  string `json:"password" validate:"required"`
	CompanyID string `json:"companyId" validate:"required"`
}

type DiscoverCompaniesRequest struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
	Name     string `json:"name,omitempty"`
}

type ConnectConnectionRequest struct {
	Login     string `json:"login" validate:"required"`
	Password  string `json:"password" validate:"required"`
	CompanyID string `json:"companyId" validate:"required"`
}

type UpdateConnectionRequest struct {
	Title  *string `json:"title,omitempty"`
	Status *string `json:"status,omitempty" validate:"omitempty,oneof=active invalid revoked sync_error"`
}

type AutoMatchRequest struct {
	Strategy string `json:"strategy" validate:"required,oneof=email"`
}

type CreateMappingRequest struct {
	EmployeeUserID uuid.UUID `json:"employeeUserId" validate:"required"`
	YougileUserID  string    `json:"yougileUserId" validate:"required"`
}

type SyncRequest struct {
	Mode             string         `json:"mode" validate:"required,oneof=incremental full"`
	IncludeUsers     bool           `json:"includeUsers"`
	IncludeStructure bool           `json:"includeStructure"`
	IncludeTasks     bool           `json:"includeTasks"`
	TaskFilters      map[string]any `json:"taskFilters,omitempty"`
}

type BackfillRequest struct {
	From      string      `json:"from" validate:"required"`
	To        string      `json:"to" validate:"required"`
	Employees []uuid.UUID `json:"employees,omitempty"`
}

type TestConnectionResponse struct {
	OK                bool   `json:"ok"`
	CompanyAccessible bool   `json:"companyAccessible"`
	RateLimitMode     string `json:"rateLimitMode"`
	Message           string `json:"message"`
}

type CreateKeyResponse struct {
	CompanyID string `json:"companyId"`
	APIKey    string `json:"apiKey"`
	Warning   string `json:"warning"`
}

type DiscoverCompaniesPaging struct {
	Count  int  `json:"count"`
	Limit  int  `json:"limit"`
	Offset int  `json:"offset"`
	Next   bool `json:"next"`
}

type DiscoverCompaniesCompany struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	IsAdmin bool   `json:"isAdmin"`
}

type DiscoverCompaniesResponse struct {
	Paging  DiscoverCompaniesPaging    `json:"paging"`
	Content []DiscoverCompaniesCompany `json:"content"`
}

type CompanyDetails struct {
	Deleted   bool           `json:"deleted"`
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	Timestamp int64          `json:"timestamp"`
	APIData   map[string]any `json:"apiData,omitempty"`
}

type ImportUsersResponse struct {
	Imported int `json:"imported"`
	Updated  int `json:"updated"`
	Failed   int `json:"failed"`
}

type ImportStructureResponse struct {
	ProjectsImported int `json:"projectsImported"`
	BoardsImported   int `json:"boardsImported"`
	ColumnsImported  int `json:"columnsImported"`
}

type AutoMatchResponse struct {
	Matched           int `json:"matched"`
	UnmatchedInternal int `json:"unmatchedInternal"`
	UnmatchedYougile  int `json:"unmatchedYougile"`
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

func (r *Repository) CreateConnection(ctx context.Context, item storedConnection, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into integration_yougile_connections (
			id, company_id, title, api_base_url, yougile_company_id, api_key_encrypted, api_key_last4,
			status, created_by, last_sync_at, last_success_sync_at, last_error, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, item.ID, item.CompanyID, item.Title, item.APIBaseURL, item.YougileCompanyID, item.APIKeyEncrypted, item.APIKeyLast4,
		item.Status, item.CreatedBy, item.LastSyncAt, item.LastSuccessSyncAt, item.LastError, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repository) UpdateConnection(ctx context.Context, item storedConnection, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update integration_yougile_connections
		set title = $2,
		    api_base_url = $3,
		    yougile_company_id = $4,
		    api_key_encrypted = $5,
		    api_key_last4 = $6,
		    status = $7,
		    last_sync_at = $8,
		    last_success_sync_at = $9,
		    last_error = $10,
		    updated_at = $11
		where id = $1
	`, item.ID, item.Title, item.APIBaseURL, item.YougileCompanyID, item.APIKeyEncrypted, item.APIKeyLast4, item.Status,
		item.LastSyncAt, item.LastSuccessSyncAt, item.LastError, item.UpdatedAt)
	return err
}

func (r *Repository) RevokeConnection(ctx context.Context, id uuid.UUID, updatedAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update integration_yougile_connections
		set status = 'revoked', updated_at = $2
		where id = $1
	`, id, updatedAt)
	return err
}

func (r *Repository) GetConnection(ctx context.Context, id uuid.UUID, exec ...db.DBTX) (storedConnection, error) {
	var item storedConnection
	err := r.base(exec...).QueryRowContext(ctx, `
			select id, company_id, title, api_base_url, yougile_company_id, api_key_encrypted, api_key_last4,
			       status, created_by, last_sync_at, last_success_sync_at, last_error, created_at, updated_at
			from integration_yougile_connections
			where id = $1
		`, id).Scan(&item.ID, &item.CompanyID, &item.Title, &item.APIBaseURL, &item.YougileCompanyID, &item.APIKeyEncrypted,
		&item.APIKeyLast4, &item.Status, &item.CreatedBy, &item.LastSyncAt, &item.LastSuccessSyncAt, &item.LastError, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repository) GetConnectionForUser(ctx context.Context, id, userID uuid.UUID, exec ...db.DBTX) (storedConnection, error) {
	var item storedConnection
	err := r.base(exec...).QueryRowContext(ctx, `
			select id, company_id, title, api_base_url, yougile_company_id, api_key_encrypted, api_key_last4,
			       status, created_by, last_sync_at, last_success_sync_at, last_error, created_at, updated_at
			from integration_yougile_connections
			where id = $1 and created_by = $2
		`, id, userID).Scan(&item.ID, &item.CompanyID, &item.Title, &item.APIBaseURL, &item.YougileCompanyID, &item.APIKeyEncrypted,
		&item.APIKeyLast4, &item.Status, &item.CreatedBy, &item.LastSyncAt, &item.LastSuccessSyncAt, &item.LastError, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repository) GetCurrentConnectionByUser(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) (storedConnection, error) {
	var item storedConnection
	err := r.base(exec...).QueryRowContext(ctx, `
			select id, company_id, title, api_base_url, yougile_company_id, api_key_encrypted, api_key_last4,
			       status, created_by, last_sync_at, last_success_sync_at, last_error, created_at, updated_at
			from integration_yougile_connections
			where created_by = $1 and status <> 'revoked'
			order by updated_at desc, created_at desc
			limit 1
		`, userID).Scan(&item.ID, &item.CompanyID, &item.Title, &item.APIBaseURL, &item.YougileCompanyID, &item.APIKeyEncrypted,
		&item.APIKeyLast4, &item.Status, &item.CreatedBy, &item.LastSyncAt, &item.LastSuccessSyncAt, &item.LastError, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repository) ListConnections(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) ([]Connection, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
			select id, company_id, title, api_base_url, yougile_company_id, api_key_last4,
			       status, created_by, last_sync_at, last_success_sync_at, last_error, created_at, updated_at
			from integration_yougile_connections
			where created_by = $1
			order by updated_at desc, created_at desc
		`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Connection
	for rows.Next() {
		var item Connection
		if err := rows.Scan(&item.ID, &item.CompanyID, &item.Title, &item.APIBaseURL, &item.YougileCompanyID, &item.APIKeyLast4,
			&item.Status, &item.CreatedBy, &item.LastSyncAt, &item.LastSuccessSyncAt, &item.LastError, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ResetConnectionData(ctx context.Context, connectionID uuid.UUID, exec ...db.DBTX) error {
	statements := []string{
		`delete from yougile_employee_metrics_daily where connection_id = $1`,
		`delete from yougile_webhook_events where connection_id = $1`,
		`delete from yougile_webhook_subscriptions where connection_id = $1`,
		`delete from yougile_sync_jobs where connection_id = $1`,
		`delete from yougile_employee_mappings where connection_id = $1`,
		`delete from yougile_users where connection_id = $1`,
		`delete from yougile_tasks where connection_id = $1`,
		`delete from yougile_columns where connection_id = $1`,
		`delete from yougile_boards where connection_id = $1`,
		`delete from yougile_projects where connection_id = $1`,
	}
	for _, statement := range statements {
		if _, err := r.base(exec...).ExecContext(ctx, statement, connectionID); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) UpsertUser(ctx context.Context, item ImportedUser, rawPayload []byte, exec ...db.DBTX) (bool, error) {
	res, err := r.base(exec...).ExecContext(ctx, `
		insert into yougile_users (
			id, connection_id, yougile_user_id, email, real_name, is_admin, status, last_activity_at, raw_payload, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10, $11)
		on conflict (connection_id, yougile_user_id) do update
		set email = excluded.email,
		    real_name = excluded.real_name,
		    is_admin = excluded.is_admin,
		    status = excluded.status,
		    last_activity_at = excluded.last_activity_at,
		    raw_payload = excluded.raw_payload,
		    updated_at = excluded.updated_at
	`, item.ID, item.ConnectionID, item.YougileUserID, item.Email, item.RealName, item.IsAdmin, item.Status, item.LastActivityAt, rawPayload, item.CreatedAt, item.UpdatedAt)
	if err != nil {
		return false, err
	}
	affected, _ := res.RowsAffected()
	return affected == 1, nil
}

func (r *Repository) ListUsers(ctx context.Context, connectionID uuid.UUID, q, email string, mapped *bool, limit, offset int) ([]ImportedUser, error) {
	query := `
			select u.id, u.connection_id, u.yougile_user_id, u.email, u.real_name, u.is_admin, u.status, u.last_activity_at, u.created_at, u.updated_at
			from yougile_users u
		`
	args := []any{connectionID}
	conds := []string{"u.connection_id = $1"}
	if mapped != nil {
		query += ` left join yougile_employee_mappings m on m.connection_id = u.connection_id and m.yougile_user_id = u.yougile_user_id and m.is_active = true `
		if *mapped {
			conds = append(conds, "m.id is not null")
		} else {
			conds = append(conds, "m.id is null")
		}
	}
	if strings.TrimSpace(q) != "" {
		args = append(args, "%"+strings.ToLower(strings.TrimSpace(q))+"%")
		conds = append(conds, fmt.Sprintf("(lower(coalesce(u.real_name, '')) like $%d or lower(coalesce(u.email, '')) like $%d)", len(args), len(args)))
	}
	if strings.TrimSpace(email) != "" {
		args = append(args, strings.ToLower(strings.TrimSpace(email)))
		conds = append(conds, fmt.Sprintf("lower(coalesce(u.email, '')) = $%d", len(args)))
	}
	args = append(args, limit, offset)
	query += " where " + strings.Join(conds, " and ") + fmt.Sprintf(" order by u.real_name asc nulls last, u.email asc nulls last limit $%d offset $%d", len(args)-1, len(args))
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ImportedUser
	for rows.Next() {
		var item ImportedUser
		if err := rows.Scan(&item.ID, &item.ConnectionID, &item.YougileUserID, &item.Email, &item.RealName, &item.IsAdmin, &item.Status, &item.LastActivityAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) UpsertProject(ctx context.Context, item Project, rawPayload []byte, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into yougile_projects (id, connection_id, yougile_project_id, title, deleted, raw_payload, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6::jsonb, $7, $8)
		on conflict (connection_id, yougile_project_id) do update
		set title = excluded.title,
		    deleted = excluded.deleted,
		    raw_payload = excluded.raw_payload,
		    updated_at = excluded.updated_at
	`, item.ID, item.ConnectionID, item.YougileProjectID, item.Title, item.Deleted, rawPayload, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repository) UpsertBoard(ctx context.Context, item Board, rawPayload []byte, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into yougile_boards (id, connection_id, yougile_board_id, yougile_project_id, title, deleted, raw_payload, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9)
		on conflict (connection_id, yougile_board_id) do update
		set yougile_project_id = excluded.yougile_project_id,
		    title = excluded.title,
		    deleted = excluded.deleted,
		    raw_payload = excluded.raw_payload,
		    updated_at = excluded.updated_at
	`, item.ID, item.ConnectionID, item.YougileBoardID, item.YougileProjectID, item.Title, item.Deleted, rawPayload, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repository) UpsertColumn(ctx context.Context, item Column, rawPayload []byte, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into yougile_columns (id, connection_id, yougile_column_id, yougile_board_id, title, color, deleted, raw_payload, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9, $10)
		on conflict (connection_id, yougile_column_id) do update
		set yougile_board_id = excluded.yougile_board_id,
		    title = excluded.title,
		    color = excluded.color,
		    deleted = excluded.deleted,
		    raw_payload = excluded.raw_payload,
		    updated_at = excluded.updated_at
	`, item.ID, item.ConnectionID, item.YougileColumnID, item.YougileBoardID, item.Title, item.Color, item.Deleted, rawPayload, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repository) ListProjects(ctx context.Context, connectionID uuid.UUID) ([]Project, error) {
	rows, err := r.db.QueryContext(ctx, `
			select id, connection_id, yougile_project_id, title, deleted, created_at, updated_at
			from yougile_projects where connection_id = $1 order by title asc
		`, connectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Project
	for rows.Next() {
		var item Project
		if err := rows.Scan(&item.ID, &item.ConnectionID, &item.YougileProjectID, &item.Title, &item.Deleted, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ListBoards(ctx context.Context, connectionID uuid.UUID) ([]Board, error) {
	rows, err := r.db.QueryContext(ctx, `
			select id, connection_id, yougile_board_id, yougile_project_id, title, deleted, created_at, updated_at
			from yougile_boards where connection_id = $1 order by title asc
		`, connectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Board
	for rows.Next() {
		var item Board
		if err := rows.Scan(&item.ID, &item.ConnectionID, &item.YougileBoardID, &item.YougileProjectID, &item.Title, &item.Deleted, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ListColumns(ctx context.Context, connectionID uuid.UUID) ([]Column, error) {
	rows, err := r.db.QueryContext(ctx, `
			select id, connection_id, yougile_column_id, yougile_board_id, title, color, deleted, created_at, updated_at
			from yougile_columns where connection_id = $1 order by title asc
		`, connectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Column
	for rows.Next() {
		var item Column
		if err := rows.Scan(&item.ID, &item.ConnectionID, &item.YougileColumnID, &item.YougileBoardID, &item.Title, &item.Color, &item.Deleted, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateOrUpdateMapping(ctx context.Context, connectionID, employeeUserID uuid.UUID, yougileUserID, source string, exec ...db.DBTX) error {
	now := time.Now().UTC()
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into yougile_employee_mappings (
			id, connection_id, employee_user_id, yougile_user_id, match_source, is_active, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, true, $6, $6)
		on conflict (connection_id, employee_user_id) do update
		set yougile_user_id = excluded.yougile_user_id,
		    match_source = excluded.match_source,
		    is_active = true,
		    updated_at = excluded.updated_at
	`, uuid.New(), connectionID, employeeUserID, yougileUserID, source, now)
	return err
}

func (r *Repository) DeleteMapping(ctx context.Context, connectionID, mappingID uuid.UUID, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		delete from yougile_employee_mappings
		where connection_id = $1 and id = $2
	`, connectionID, mappingID)
	return err
}

func (r *Repository) ListMappings(ctx context.Context, connectionID uuid.UUID) ([]MappingView, error) {
	rows, err := r.db.QueryContext(ctx, `
			select m.id,
			       m.employee_user_id,
			       concat_ws(' ', ep.last_name, ep.first_name, ep.middle_name) as employee_name,
			       u.email,
			       m.yougile_user_id,
			       yu.real_name,
			       yu.email,
			       m.match_source,
			       m.is_active,
			       m.created_at,
			       m.updated_at
			from yougile_employee_mappings m
			join users u on u.id = m.employee_user_id
			left join employee_profiles ep on ep.user_id = m.employee_user_id
			left join yougile_users yu on yu.connection_id = m.connection_id and yu.yougile_user_id = m.yougile_user_id
			where m.connection_id = $1
			order by employee_name asc, u.email asc
		`, connectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []MappingView
	for rows.Next() {
		var item MappingView
		if err := rows.Scan(&item.ID, &item.EmployeeUserID, &item.EmployeeName, &item.EmployeeEmail, &item.YougileUserID,
			&item.YougileRealName, &item.YougileEmail, &item.MatchSource, &item.IsActive, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ListInternalUsersForMatching(ctx context.Context) ([]struct {
	UserID uuid.UUID
	Email  string
}, error) {
	rows, err := r.db.QueryContext(ctx, `select id, lower(email) from users where deleted_at is null`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []struct {
		UserID uuid.UUID
		Email  string
	}
	for rows.Next() {
		var item struct {
			UserID uuid.UUID
			Email  string
		}
		if err := rows.Scan(&item.UserID, &item.Email); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ListImportedUsersForMatching(ctx context.Context, connectionID uuid.UUID) ([]ImportedUser, error) {
	return r.ListUsers(ctx, connectionID, "", "", nil, 10000, 0)
}

func (r *Repository) CreateSyncJob(ctx context.Context, item SyncJob, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into yougile_sync_jobs (
			id, connection_id, job_type, status, cursor, progress, started_at, finished_at,
			attempt, next_retry_at, error_text, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7, $8, $9, $10, $11, $12, $13)
	`, item.ID, item.ConnectionID, item.JobType, item.Status, nullableJSON(item.Cursor), nullableJSON(item.Progress),
		item.StartedAt, item.FinishedAt, item.Attempt, item.NextRetryAt, item.ErrorText, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repository) GetSyncJob(ctx context.Context, id uuid.UUID, exec ...db.DBTX) (SyncJob, error) {
	var item SyncJob
	err := r.base(exec...).QueryRowContext(ctx, `
			select id, connection_id, job_type, status, coalesce(cursor, '{}'::jsonb)::text, coalesce(progress, '{}'::jsonb)::text,
			       started_at, finished_at, attempt, next_retry_at, error_text, created_at, updated_at
			from yougile_sync_jobs where id = $1
		`, id).Scan(&item.ID, &item.ConnectionID, &item.JobType, &item.Status, &item.Cursor, &item.Progress, &item.StartedAt, &item.FinishedAt,
		&item.Attempt, &item.NextRetryAt, &item.ErrorText, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repository) GetSyncJobForUser(ctx context.Context, id, userID uuid.UUID, exec ...db.DBTX) (SyncJob, error) {
	var item SyncJob
	err := r.base(exec...).QueryRowContext(ctx, `
			select job.id, job.connection_id, job.job_type, job.status,
			       coalesce(job.cursor, '{}'::jsonb)::text, coalesce(job.progress, '{}'::jsonb)::text,
			       job.started_at, job.finished_at, job.attempt, job.next_retry_at, job.error_text, job.created_at, job.updated_at
			from yougile_sync_jobs job
			join integration_yougile_connections connection on connection.id = job.connection_id
			where job.id = $1 and connection.created_by = $2
		`, id, userID).Scan(&item.ID, &item.ConnectionID, &item.JobType, &item.Status, &item.Cursor, &item.Progress, &item.StartedAt, &item.FinishedAt,
		&item.Attempt, &item.NextRetryAt, &item.ErrorText, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repository) MarkSyncJobProcessing(ctx context.Context, id uuid.UUID, progress []byte, startedAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update yougile_sync_jobs
		set status = 'processing', started_at = coalesce(started_at, $2), attempt = attempt + 1, progress = $3::jsonb, updated_at = $2
		where id = $1
	`, id, startedAt, progress)
	return err
}

func (r *Repository) MarkSyncJobDone(ctx context.Context, id uuid.UUID, progress []byte, finishedAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update yougile_sync_jobs
		set status = 'done', progress = $3::jsonb, finished_at = $2, error_text = null, updated_at = $2
		where id = $1
	`, id, finishedAt, progress)
	return err
}

func (r *Repository) MarkSyncJobFailed(ctx context.Context, id uuid.UUID, progress []byte, finishedAt time.Time, message string, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update yougile_sync_jobs
		set status = 'failed', progress = $3::jsonb, finished_at = $2, error_text = $4, updated_at = $2
		where id = $1
	`, id, finishedAt, progress, message)
	return err
}

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "https://yougile.com"
	}
	return &Client{
		baseURL:    baseURL,
		apiKey:     strings.TrimSpace(apiKey),
		httpClient: &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *Client) DiscoverCompanies(ctx context.Context, login, password, name string) (DiscoverCompaniesResponse, error) {
	body := map[string]any{
		"login":    login,
		"password": password,
	}
	if trimmedName := strings.TrimSpace(name); trimmedName != "" {
		body["name"] = trimmedName
	}
	payload, err := c.doJSON(ctx, http.MethodPost, "/api-v2/auth/companies?limit=1000", body, false)
	if err != nil {
		return DiscoverCompaniesResponse{}, err
	}
	result := DiscoverCompaniesResponse{
		Content: []DiscoverCompaniesCompany{},
	}
	if paging, ok := payload["paging"].(map[string]any); ok {
		result.Paging = DiscoverCompaniesPaging{
			Count:  intFromAny(paging["count"]),
			Limit:  intFromAny(paging["limit"]),
			Offset: intFromAny(paging["offset"]),
			Next:   boolFromAny(paging["next"]),
		}
	}
	for _, item := range extractItems(payload) {
		result.Content = append(result.Content, DiscoverCompaniesCompany{
			ID:      firstString(item, "id"),
			Name:    firstString(item, "name", "title"),
			IsAdmin: boolFromAny(item["isAdmin"]),
		})
	}
	return result, nil
}

func (c *Client) CreateKey(ctx context.Context, login, password, companyID string) (string, error) {
	payload, err := c.doJSON(ctx, http.MethodPost, "/api-v2/auth/keys", map[string]any{
		"login":     login,
		"password":  password,
		"companyId": companyID,
	}, false)
	if err != nil {
		return "", err
	}
	key := stringFromAny(payload["key"])
	if key == "" {
		key = stringFromAny(payload["apiKey"])
	}
	if key == "" {
		return "", httpx.Internal("yougile_invalid_response")
	}
	return key, nil
}

func (c *Client) GetCompany(ctx context.Context, companyID string) (CompanyDetails, error) {
	payload, err := c.doJSON(ctx, http.MethodGet, "/api-v2/companies/"+url.PathEscape(strings.TrimSpace(companyID)), nil, true)
	if err != nil {
		return CompanyDetails{}, err
	}

	result := CompanyDetails{
		Deleted:   boolFromAny(payload["deleted"]),
		ID:        firstString(payload, "id"),
		Title:     firstString(payload, "title", "name"),
		Timestamp: int64(intFromAny(payload["timestamp"])),
	}
	if apiData, ok := payload["apiData"].(map[string]any); ok {
		result.APIData = apiData
	}
	return result, nil
}

func (c *Client) TestKey(ctx context.Context) error {
	_, err := c.doJSON(ctx, http.MethodGet, "/api-v2/users?limit=1", nil, true)
	return err
}

func (c *Client) ListUsers(ctx context.Context) ([]map[string]any, error) {
	payload, err := c.doJSON(ctx, http.MethodGet, "/api-v2/users?limit=1000", nil, true)
	if err != nil {
		return nil, err
	}
	return extractItems(payload), nil
}

func (c *Client) ListProjects(ctx context.Context) ([]map[string]any, error) {
	payload, err := c.doJSON(ctx, http.MethodGet, "/api-v2/projects", nil, true)
	if err != nil {
		return nil, err
	}
	return extractItems(payload), nil
}

func (c *Client) ListBoards(ctx context.Context) ([]map[string]any, error) {
	payload, err := c.doJSON(ctx, http.MethodGet, "/api-v2/boards", nil, true)
	if err != nil {
		return nil, err
	}
	return extractItems(payload), nil
}

func (c *Client) ListColumns(ctx context.Context) ([]map[string]any, error) {
	payload, err := c.doJSON(ctx, http.MethodGet, "/api-v2/columns", nil, true)
	if err != nil {
		return nil, err
	}
	return extractItems(payload), nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, auth bool) (map[string]any, error) {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if auth && c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, httpx.NewError(resp.StatusCode, "yougile_request_failed", strings.TrimSpace(string(raw)))
	}

	if len(bytes.TrimSpace(raw)) == 0 {
		return map[string]any{}, nil
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err == nil {
		return payload, nil
	}

	var arr []map[string]any
	if err := json.Unmarshal(raw, &arr); err == nil {
		return map[string]any{"content": arr}, nil
	}

	return nil, httpx.Internal("yougile_invalid_response")
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

func (s *Service) getOwnedConnection(ctx context.Context, userID, connectionID uuid.UUID) (storedConnection, error) {
	item, err := s.repo.GetConnectionForUser(ctx, connectionID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storedConnection{}, httpx.NotFound("yougile_connection_not_found", "yougile connection not found")
		}
		return storedConnection{}, err
	}
	if item.Status == "revoked" {
		return storedConnection{}, httpx.NotFound("yougile_connection_not_found", "yougile connection not found")
	}
	return item, nil
}

func (s *Service) saveConnection(ctx context.Context, principal platformauth.Principal, title *string, apiBaseURL, companyID, apiKey string) (Connection, error) {
	companyID = strings.TrimSpace(companyID)
	apiKey = strings.TrimSpace(apiKey)
	now := s.clock.Now()
	last4 := lastN(apiKey, 4)
	baseURL := defaultBaseURL(apiBaseURL)

	var saved Connection
	err := db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		current, err := s.repo.GetCurrentConnectionByUser(ctx, principal.UserID, tx)
		switch {
		case err == nil:
			if current.YougileCompanyID != companyID {
				if err := s.repo.ResetConnectionData(ctx, current.ID, tx); err != nil {
					return err
				}
				current.LastSyncAt = nil
				current.LastSuccessSyncAt = nil
			}
			current.Title = trimStringPtr(title)
			current.APIBaseURL = baseURL
			current.YougileCompanyID = companyID
			current.APIKeyEncrypted = apiKey
			current.APIKeyLast4 = stringPtr(last4)
			current.Status = "active"
			current.LastError = nil
			current.UpdatedAt = now
			if err := s.repo.UpdateConnection(ctx, current, tx); err != nil {
				return err
			}
			saved = current.Connection
			return nil
		case !errors.Is(err, sql.ErrNoRows):
			return err
		}

		item := storedConnection{
			Connection: Connection{
				ID:               uuid.New(),
				Title:            trimStringPtr(title),
				APIBaseURL:       baseURL,
				YougileCompanyID: companyID,
				APIKeyLast4:      stringPtr(last4),
				Status:           "active",
				CreatedBy:        principal.UserID,
				CreatedAt:        now,
				UpdatedAt:        now,
			},
			APIKeyEncrypted: apiKey,
		}
		if err := s.repo.CreateConnection(ctx, item, tx); err != nil {
			return err
		}
		saved = item.Connection
		return nil
	})
	if err != nil {
		return Connection{}, err
	}
	return saved, nil
}

func (s *Service) CreateConnection(ctx context.Context, principal platformauth.Principal, req CreateConnectionRequest) (Connection, error) {
	client := NewClient(req.APIBaseURL, req.APIKey)
	if err := client.TestKey(ctx); err != nil {
		return Connection{}, err
	}
	return s.saveConnection(ctx, principal, req.Title, req.APIBaseURL, req.YougileCompanyID, req.APIKey)
}

func (s *Service) ConnectConnection(ctx context.Context, principal platformauth.Principal, req ConnectConnectionRequest) (Connection, error) {
	client := NewClient("", "")
	key, err := client.CreateKey(ctx, req.Login, req.Password, req.CompanyID)
	if err != nil {
		return Connection{}, err
	}

	details, err := NewClient("", key).GetCompany(ctx, req.CompanyID)
	if err != nil {
		return Connection{}, err
	}

	title := trimStringPtr(stringPtr(details.Title))
	companyID := strings.TrimSpace(details.ID)
	if companyID == "" {
		companyID = strings.TrimSpace(req.CompanyID)
	}

	return s.saveConnection(ctx, principal, title, "", companyID, key)
}

func (s *Service) TestKey(ctx context.Context, req TestKeyRequest) (TestConnectionResponse, error) {
	client := NewClient(req.APIBaseURL, req.APIKey)
	if err := client.TestKey(ctx); err != nil {
		return TestConnectionResponse{}, err
	}
	return TestConnectionResponse{
		OK:                true,
		CompanyAccessible: true,
		RateLimitMode:     "50_per_minute",
		Message:           "Connection verified",
	}, nil
}

func (s *Service) CreateKey(ctx context.Context, req CreateKeyRequest) (CreateKeyResponse, error) {
	client := NewClient("", "")
	key, err := client.CreateKey(ctx, req.Login, req.Password, req.CompanyID)
	if err != nil {
		return CreateKeyResponse{}, err
	}
	return CreateKeyResponse{
		CompanyID: req.CompanyID,
		APIKey:    key,
		Warning:   "Store securely",
	}, nil
}

func (s *Service) DiscoverCompanies(ctx context.Context, req DiscoverCompaniesRequest) (DiscoverCompaniesResponse, error) {
	return NewClient("", "").DiscoverCompanies(ctx, req.Login, req.Password, req.Name)
}

func (s *Service) ListConnections(ctx context.Context, principal platformauth.Principal) ([]Connection, error) {
	return s.repo.ListConnections(ctx, principal.UserID)
}

func (s *Service) GetConnection(ctx context.Context, principal platformauth.Principal, id uuid.UUID) (Connection, error) {
	item, err := s.getOwnedConnection(ctx, principal.UserID, id)
	if err != nil {
		return Connection{}, err
	}
	return item.Connection, nil
}

func (s *Service) UpdateConnection(ctx context.Context, principal platformauth.Principal, id uuid.UUID, req UpdateConnectionRequest) (Connection, error) {
	item, err := s.getOwnedConnection(ctx, principal.UserID, id)
	if err != nil {
		return Connection{}, err
	}
	if req.Title != nil {
		item.Title = trimStringPtr(req.Title)
	}
	if req.Status != nil {
		item.Status = strings.TrimSpace(*req.Status)
	}
	item.UpdatedAt = s.clock.Now()
	if err := s.repo.UpdateConnection(ctx, item); err != nil {
		return Connection{}, err
	}
	return item.Connection, nil
}

func (s *Service) DeleteConnection(ctx context.Context, principal platformauth.Principal, id uuid.UUID) error {
	if _, err := s.getOwnedConnection(ctx, principal.UserID, id); err != nil {
		return err
	}
	return s.repo.RevokeConnection(ctx, id, s.clock.Now())
}

func (s *Service) ImportUsers(ctx context.Context, connectionID uuid.UUID) (ImportUsersResponse, error) {
	conn, err := s.repo.GetConnection(ctx, connectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ImportUsersResponse{}, httpx.NotFound("yougile_connection_not_found", "yougile connection not found")
		}
		return ImportUsersResponse{}, err
	}
	users, err := NewClient(conn.APIBaseURL, conn.APIKeyEncrypted).ListUsers(ctx)
	if err != nil {
		return ImportUsersResponse{}, err
	}
	now := s.clock.Now()
	result := ImportUsersResponse{}
	for _, raw := range users {
		item := ImportedUser{
			ID:            uuid.New(),
			ConnectionID:  connectionID,
			YougileUserID: firstString(raw, "id", "_id", "userId"),
			Email:         normalizeOptionalString(raw["email"]),
			RealName:      normalizeOptionalString(raw["realName"], raw["name"], raw["fullName"]),
			IsAdmin:       boolFromAny(raw["isAdmin"]),
			Status:        normalizeOptionalString(raw["status"]),
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if item.YougileUserID == "" {
			result.Failed++
			continue
		}
		if ts := timeFromAny(raw["lastActivityAt"]); ts != nil {
			item.LastActivityAt = ts
		}
		payload, _ := json.Marshal(raw)
		inserted, err := s.repo.UpsertUser(ctx, item, payload)
		if err != nil {
			result.Failed++
			continue
		}
		if inserted {
			result.Imported++
		} else {
			result.Updated++
		}
	}
	conn.LastSyncAt = &now
	conn.LastSuccessSyncAt = &now
	conn.LastError = nil
	conn.UpdatedAt = now
	_ = s.repo.UpdateConnection(ctx, conn)
	return result, nil
}

func (s *Service) ImportStructure(ctx context.Context, connectionID uuid.UUID) (ImportStructureResponse, error) {
	conn, err := s.repo.GetConnection(ctx, connectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ImportStructureResponse{}, httpx.NotFound("yougile_connection_not_found", "yougile connection not found")
		}
		return ImportStructureResponse{}, err
	}
	client := NewClient(conn.APIBaseURL, conn.APIKeyEncrypted)
	projects, err := client.ListProjects(ctx)
	if err != nil {
		return ImportStructureResponse{}, err
	}
	boards, err := client.ListBoards(ctx)
	if err != nil {
		return ImportStructureResponse{}, err
	}
	columns, err := client.ListColumns(ctx)
	if err != nil {
		return ImportStructureResponse{}, err
	}
	now := s.clock.Now()
	result := ImportStructureResponse{}
	for _, raw := range projects {
		item := Project{
			ID:               uuid.New(),
			ConnectionID:     connectionID,
			YougileProjectID: firstString(raw, "id", "_id", "projectId"),
			Title:            firstString(raw, "title", "name"),
			Deleted:          boolFromAny(raw["deleted"]),
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if item.YougileProjectID == "" || strings.TrimSpace(item.Title) == "" {
			continue
		}
		payload, _ := json.Marshal(raw)
		if err := s.repo.UpsertProject(ctx, item, payload); err == nil {
			result.ProjectsImported++
		}
	}
	for _, raw := range boards {
		item := Board{
			ID:               uuid.New(),
			ConnectionID:     connectionID,
			YougileBoardID:   firstString(raw, "id", "_id", "boardId"),
			YougileProjectID: firstString(raw, "projectId"),
			Title:            firstString(raw, "title", "name"),
			Deleted:          boolFromAny(raw["deleted"]),
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if item.YougileBoardID == "" || strings.TrimSpace(item.Title) == "" {
			continue
		}
		payload, _ := json.Marshal(raw)
		if err := s.repo.UpsertBoard(ctx, item, payload); err == nil {
			result.BoardsImported++
		}
	}
	for _, raw := range columns {
		item := Column{
			ID:              uuid.New(),
			ConnectionID:    connectionID,
			YougileColumnID: firstString(raw, "id", "_id", "columnId"),
			YougileBoardID:  firstString(raw, "boardId"),
			Title:           firstString(raw, "title", "name"),
			Color:           intPtrFromAny(raw["color"]),
			Deleted:         boolFromAny(raw["deleted"]),
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if item.YougileColumnID == "" || strings.TrimSpace(item.Title) == "" {
			continue
		}
		payload, _ := json.Marshal(raw)
		if err := s.repo.UpsertColumn(ctx, item, payload); err == nil {
			result.ColumnsImported++
		}
	}
	conn.LastSyncAt = &now
	conn.LastSuccessSyncAt = &now
	conn.LastError = nil
	conn.UpdatedAt = now
	_ = s.repo.UpdateConnection(ctx, conn)
	return result, nil
}

func (s *Service) ListUsers(ctx context.Context, connectionID uuid.UUID, q, email string, mapped *bool, limit, offset int) ([]ImportedUser, error) {
	return s.repo.ListUsers(ctx, connectionID, q, email, mapped, limit, offset)
}

func (s *Service) ListProjects(ctx context.Context, connectionID uuid.UUID) ([]Project, error) {
	return s.repo.ListProjects(ctx, connectionID)
}

func (s *Service) ListBoards(ctx context.Context, connectionID uuid.UUID) ([]Board, error) {
	return s.repo.ListBoards(ctx, connectionID)
}

func (s *Service) ListColumns(ctx context.Context, connectionID uuid.UUID) ([]Column, error) {
	return s.repo.ListColumns(ctx, connectionID)
}

func (s *Service) AutoMatch(ctx context.Context, connectionID uuid.UUID, strategy string) (AutoMatchResponse, error) {
	if strategy != "email" {
		return AutoMatchResponse{}, httpx.BadRequest("invalid_strategy", "only email strategy is supported")
	}
	internalUsers, err := s.repo.ListInternalUsersForMatching(ctx)
	if err != nil {
		return AutoMatchResponse{}, err
	}
	importedUsers, err := s.repo.ListImportedUsersForMatching(ctx, connectionID)
	if err != nil {
		return AutoMatchResponse{}, err
	}
	internalByEmail := make(map[string]uuid.UUID, len(internalUsers))
	for _, item := range internalUsers {
		internalByEmail[strings.ToLower(strings.TrimSpace(item.Email))] = item.UserID
	}
	matchedInternal := make(map[uuid.UUID]struct{})
	matchedYougile := make(map[string]struct{})
	result := AutoMatchResponse{}
	for _, yu := range importedUsers {
		if yu.Email == nil || strings.TrimSpace(*yu.Email) == "" {
			continue
		}
		userID, ok := internalByEmail[strings.ToLower(strings.TrimSpace(*yu.Email))]
		if !ok {
			continue
		}
		if err := s.repo.CreateOrUpdateMapping(ctx, connectionID, userID, yu.YougileUserID, "email"); err != nil {
			return AutoMatchResponse{}, err
		}
		matchedInternal[userID] = struct{}{}
		matchedYougile[yu.YougileUserID] = struct{}{}
		result.Matched++
	}
	result.UnmatchedInternal = len(internalUsers) - len(matchedInternal)
	result.UnmatchedYougile = len(importedUsers) - len(matchedYougile)
	if result.UnmatchedInternal < 0 {
		result.UnmatchedInternal = 0
	}
	if result.UnmatchedYougile < 0 {
		result.UnmatchedYougile = 0
	}
	return result, nil
}

func (s *Service) ListMappings(ctx context.Context, connectionID uuid.UUID) ([]MappingView, error) {
	return s.repo.ListMappings(ctx, connectionID)
}

func (s *Service) CreateMapping(ctx context.Context, connectionID uuid.UUID, req CreateMappingRequest) error {
	return s.repo.CreateOrUpdateMapping(ctx, connectionID, req.EmployeeUserID, strings.TrimSpace(req.YougileUserID), "manual")
}

func (s *Service) DeleteMapping(ctx context.Context, connectionID, mappingID uuid.UUID) error {
	return s.repo.DeleteMapping(ctx, connectionID, mappingID)
}

func (s *Service) StartSync(ctx context.Context, connectionID uuid.UUID, req SyncRequest) (SyncJob, error) {
	now := s.clock.Now()
	jobType := "full_sync"
	switch {
	case req.IncludeUsers && !req.IncludeStructure && !req.IncludeTasks:
		jobType = "users_sync"
	case !req.IncludeUsers && req.IncludeStructure && !req.IncludeTasks:
		jobType = "structure_sync"
	case !req.IncludeUsers && !req.IncludeStructure && req.IncludeTasks:
		jobType = "tasks_sync"
	}
	progress := map[string]any{
		"users":     0,
		"structure": 0,
		"tasks":     0,
		"mode":      req.Mode,
	}
	progressJSON, _ := json.Marshal(progress)
	job := SyncJob{
		ID:           uuid.New(),
		ConnectionID: connectionID,
		JobType:      jobType,
		Status:       "pending",
		Cursor:       json.RawMessage(`{}`),
		Progress:     progressJSON,
		Attempt:      0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.CreateSyncJob(ctx, job); err != nil {
		return SyncJob{}, err
	}
	payload := map[string]any{
		"sync_job_id":       job.ID,
		"connection_id":     connectionID,
		"mode":              req.Mode,
		"include_users":     req.IncludeUsers,
		"include_structure": req.IncludeStructure,
		"include_tasks":     req.IncludeTasks,
		"task_filters":      req.TaskFilters,
	}
	idempotencyKey := "yougile-sync:" + job.ID.String()
	if err := s.queue.Enqueue(ctx, s.db, "integrations", "yougile_sync", payload, &idempotencyKey, now); err != nil {
		return SyncJob{}, err
	}
	return job, nil
}

func (s *Service) StartBackfill(ctx context.Context, connectionID uuid.UUID, req BackfillRequest) (SyncJob, error) {
	if _, err := time.Parse("2006-01-02", req.From); err != nil {
		return SyncJob{}, httpx.BadRequest("invalid_from", "from must be in YYYY-MM-DD format")
	}
	if _, err := time.Parse("2006-01-02", req.To); err != nil {
		return SyncJob{}, httpx.BadRequest("invalid_to", "to must be in YYYY-MM-DD format")
	}
	now := s.clock.Now()
	progressJSON, _ := json.Marshal(map[string]any{
		"from": req.From,
		"to":   req.To,
	})
	job := SyncJob{
		ID:           uuid.New(),
		ConnectionID: connectionID,
		JobType:      "backfill",
		Status:       "pending",
		Cursor:       json.RawMessage(`{}`),
		Progress:     progressJSON,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.CreateSyncJob(ctx, job); err != nil {
		return SyncJob{}, err
	}
	payload := map[string]any{
		"sync_job_id":   job.ID,
		"connection_id": connectionID,
		"from":          req.From,
		"to":            req.To,
		"employees":     req.Employees,
	}
	idempotencyKey := "yougile-backfill:" + job.ID.String()
	if err := s.queue.Enqueue(ctx, s.db, "integrations", "yougile_sync", payload, &idempotencyKey, now); err != nil {
		return SyncJob{}, err
	}
	return job, nil
}

func (s *Service) GetSyncJob(ctx context.Context, principal platformauth.Principal, id uuid.UUID) (SyncJob, error) {
	job, err := s.repo.GetSyncJobForUser(ctx, id, principal.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SyncJob{}, httpx.NotFound("yougile_sync_job_not_found", "yougile sync job not found")
		}
		return SyncJob{}, err
	}
	return job, nil
}

func (s *Service) ProcessSyncJob(ctx context.Context, job worker.Job) error {
	var payload struct {
		SyncJobID        uuid.UUID `json:"sync_job_id"`
		ConnectionID     uuid.UUID `json:"connection_id"`
		IncludeUsers     bool      `json:"include_users"`
		IncludeStructure bool      `json:"include_structure"`
		IncludeTasks     bool      `json:"include_tasks"`
	}
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return err
	}
	now := s.clock.Now()
	progress := map[string]any{"status": "processing"}
	progressJSON, _ := json.Marshal(progress)
	if err := s.repo.MarkSyncJobProcessing(ctx, payload.SyncJobID, progressJSON, now); err != nil {
		return err
	}
	if payload.IncludeUsers {
		result, err := s.ImportUsers(ctx, payload.ConnectionID)
		if err != nil {
			progress["users_error"] = err.Error()
			progressJSON, _ = json.Marshal(progress)
			_ = s.repo.MarkSyncJobFailed(ctx, payload.SyncJobID, progressJSON, s.clock.Now(), err.Error())
			return err
		}
		progress["users"] = result
	}
	if payload.IncludeStructure {
		result, err := s.ImportStructure(ctx, payload.ConnectionID)
		if err != nil {
			progress["structure_error"] = err.Error()
			progressJSON, _ = json.Marshal(progress)
			_ = s.repo.MarkSyncJobFailed(ctx, payload.SyncJobID, progressJSON, s.clock.Now(), err.Error())
			return err
		}
		progress["structure"] = result
	}
	if payload.IncludeTasks {
		progress["tasks"] = "queued_but_not_implemented"
	}
	progress["status"] = "done"
	progressJSON, _ = json.Marshal(progress)
	return s.repo.MarkSyncJobDone(ctx, payload.SyncJobID, progressJSON, s.clock.Now())
}

type Handler struct {
	service   *Service
	validator *validator.Validate
}

func NewHandler(service *Service, validate *validator.Validate) *Handler {
	return &Handler{service: service, validator: validate}
}

func yougilePrincipal(r *http.Request) (platformauth.Principal, error) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		return platformauth.Principal{}, httpx.Unauthorized("unauthorized", "authorization required")
	}
	return principal, nil
}

func requireYougilePrincipal(w http.ResponseWriter, r *http.Request) (platformauth.Principal, bool) {
	principal, err := yougilePrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return platformauth.Principal{}, false
	}
	return principal, true
}

func (h *Handler) CreateConnection(w http.ResponseWriter, r *http.Request) {
	principal, ok := requireYougilePrincipal(w, r)
	if !ok {
		return
	}
	var req CreateConnectionRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_request", err.Error()))
		return
	}
	item, err := h.service.CreateConnection(r.Context(), principal, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) ConnectConnection(w http.ResponseWriter, r *http.Request) {
	principal, ok := requireYougilePrincipal(w, r)
	if !ok {
		return
	}
	var req ConnectConnectionRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_request", err.Error()))
		return
	}
	item, err := h.service.ConnectConnection(r.Context(), principal, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) TestKey(w http.ResponseWriter, r *http.Request) {
	var req TestKeyRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_request", err.Error()))
		return
	}
	item, err := h.service.TestKey(r.Context(), req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) CreateKey(w http.ResponseWriter, r *http.Request) {
	var req CreateKeyRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_request", err.Error()))
		return
	}
	item, err := h.service.CreateKey(r.Context(), req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) DiscoverCompanies(w http.ResponseWriter, r *http.Request) {
	var req DiscoverCompaniesRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_request", err.Error()))
		return
	}
	item, err := h.service.DiscoverCompanies(r.Context(), req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) ListConnections(w http.ResponseWriter, r *http.Request) {
	principal, ok := requireYougilePrincipal(w, r)
	if !ok {
		return
	}
	items, err := h.service.ListConnections(r.Context(), principal)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) GetConnection(w http.ResponseWriter, r *http.Request) {
	principal, ok := requireYougilePrincipal(w, r)
	if !ok {
		return
	}
	id, err := parseUUIDParam(chi.URLParam(r, "id"), "invalid_connection_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	item, err := h.service.GetConnection(r.Context(), principal, id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) UpdateConnection(w http.ResponseWriter, r *http.Request) {
	principal, ok := requireYougilePrincipal(w, r)
	if !ok {
		return
	}
	id, err := parseUUIDParam(chi.URLParam(r, "id"), "invalid_connection_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	var req UpdateConnectionRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	item, err := h.service.UpdateConnection(r.Context(), principal, id, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) DeleteConnection(w http.ResponseWriter, r *http.Request) {
	principal, ok := requireYougilePrincipal(w, r)
	if !ok {
		return
	}
	id, err := parseUUIDParam(chi.URLParam(r, "id"), "invalid_connection_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.service.DeleteConnection(r.Context(), principal, id); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) ImportUsers(w http.ResponseWriter, r *http.Request) {
	principal, ok := requireYougilePrincipal(w, r)
	if !ok {
		return
	}
	id, err := parseUUIDParam(chi.URLParam(r, "id"), "invalid_connection_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if _, err := h.service.GetConnection(r.Context(), principal, id); err != nil {
		httpx.WriteError(w, err)
		return
	}
	item, err := h.service.ImportUsers(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) ImportStructure(w http.ResponseWriter, r *http.Request) {
	principal, ok := requireYougilePrincipal(w, r)
	if !ok {
		return
	}
	id, err := parseUUIDParam(chi.URLParam(r, "id"), "invalid_connection_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if _, err := h.service.GetConnection(r.Context(), principal, id); err != nil {
		httpx.WriteError(w, err)
		return
	}
	item, err := h.service.ImportStructure(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	principal, ok := requireYougilePrincipal(w, r)
	if !ok {
		return
	}
	id, err := parseUUIDParam(chi.URLParam(r, "id"), "invalid_connection_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	limit, offset := parsePaging(r)
	var mapped *bool
	if raw := strings.TrimSpace(r.URL.Query().Get("mapped")); raw != "" {
		value := raw == "true" || raw == "1"
		mapped = &value
	}
	if _, err := h.service.GetConnection(r.Context(), principal, id); err != nil {
		httpx.WriteError(w, err)
		return
	}
	items, err := h.service.ListUsers(r.Context(), id, r.URL.Query().Get("q"), r.URL.Query().Get("email"), mapped, limit, offset)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	principal, ok := requireYougilePrincipal(w, r)
	if !ok {
		return
	}
	id, err := parseUUIDParam(chi.URLParam(r, "id"), "invalid_connection_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if _, err := h.service.GetConnection(r.Context(), principal, id); err != nil {
		httpx.WriteError(w, err)
		return
	}
	items, err := h.service.ListProjects(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) ListBoards(w http.ResponseWriter, r *http.Request) {
	principal, ok := requireYougilePrincipal(w, r)
	if !ok {
		return
	}
	id, err := parseUUIDParam(chi.URLParam(r, "id"), "invalid_connection_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if _, err := h.service.GetConnection(r.Context(), principal, id); err != nil {
		httpx.WriteError(w, err)
		return
	}
	items, err := h.service.ListBoards(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) ListColumns(w http.ResponseWriter, r *http.Request) {
	principal, ok := requireYougilePrincipal(w, r)
	if !ok {
		return
	}
	id, err := parseUUIDParam(chi.URLParam(r, "id"), "invalid_connection_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if _, err := h.service.GetConnection(r.Context(), principal, id); err != nil {
		httpx.WriteError(w, err)
		return
	}
	items, err := h.service.ListColumns(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) AutoMatch(w http.ResponseWriter, r *http.Request) {
	principal, ok := requireYougilePrincipal(w, r)
	if !ok {
		return
	}
	id, err := parseUUIDParam(chi.URLParam(r, "id"), "invalid_connection_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	var req AutoMatchRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_request", err.Error()))
		return
	}
	if _, err := h.service.GetConnection(r.Context(), principal, id); err != nil {
		httpx.WriteError(w, err)
		return
	}
	item, err := h.service.AutoMatch(r.Context(), id, req.Strategy)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) ListMappings(w http.ResponseWriter, r *http.Request) {
	principal, ok := requireYougilePrincipal(w, r)
	if !ok {
		return
	}
	id, err := parseUUIDParam(chi.URLParam(r, "id"), "invalid_connection_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if _, err := h.service.GetConnection(r.Context(), principal, id); err != nil {
		httpx.WriteError(w, err)
		return
	}
	items, err := h.service.ListMappings(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"content": items})
}

func (h *Handler) CreateMapping(w http.ResponseWriter, r *http.Request) {
	principal, ok := requireYougilePrincipal(w, r)
	if !ok {
		return
	}
	id, err := parseUUIDParam(chi.URLParam(r, "id"), "invalid_connection_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	var req CreateMappingRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_request", err.Error()))
		return
	}
	if _, err := h.service.GetConnection(r.Context(), principal, id); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.service.CreateMapping(r.Context(), id, req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) DeleteMapping(w http.ResponseWriter, r *http.Request) {
	principal, ok := requireYougilePrincipal(w, r)
	if !ok {
		return
	}
	connectionID, err := parseUUIDParam(chi.URLParam(r, "id"), "invalid_connection_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	mappingID, err := parseUUIDParam(chi.URLParam(r, "mappingId"), "invalid_mapping_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if _, err := h.service.GetConnection(r.Context(), principal, connectionID); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.service.DeleteMapping(r.Context(), connectionID, mappingID); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) StartSync(w http.ResponseWriter, r *http.Request) {
	principal, ok := requireYougilePrincipal(w, r)
	if !ok {
		return
	}
	id, err := parseUUIDParam(chi.URLParam(r, "id"), "invalid_connection_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	var req SyncRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_request", err.Error()))
		return
	}
	if _, err := h.service.GetConnection(r.Context(), principal, id); err != nil {
		httpx.WriteError(w, err)
		return
	}
	item, err := h.service.StartSync(r.Context(), id, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"jobId": item.ID, "status": item.Status})
}

func (h *Handler) GetSyncJob(w http.ResponseWriter, r *http.Request) {
	principal, ok := requireYougilePrincipal(w, r)
	if !ok {
		return
	}
	id, err := parseUUIDParam(chi.URLParam(r, "jobId"), "invalid_job_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	item, err := h.service.GetSyncJob(r.Context(), principal, id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) Backfill(w http.ResponseWriter, r *http.Request) {
	principal, ok := requireYougilePrincipal(w, r)
	if !ok {
		return
	}
	id, err := parseUUIDParam(chi.URLParam(r, "id"), "invalid_connection_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	var req BackfillRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_request", err.Error()))
		return
	}
	if _, err := h.service.GetConnection(r.Context(), principal, id); err != nil {
		httpx.WriteError(w, err)
		return
	}
	item, err := h.service.StartBackfill(r.Context(), id, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"jobId": item.ID, "status": item.Status})
}

func parseUUIDParam(raw, code string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return uuid.Nil, httpx.BadRequest(code, "invalid uuid")
	}
	return id, nil
}

func parsePaging(r *http.Request) (limit, offset int) {
	limit = 50
	offset = 0
	values := r.URL.Query()
	if raw := strings.TrimSpace(values.Get("limit")); raw != "" {
		var parsed int
		fmt.Sscanf(raw, "%d", &parsed)
		if parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}
	if raw := strings.TrimSpace(values.Get("offset")); raw != "" {
		var parsed int
		fmt.Sscanf(raw, "%d", &parsed)
		if parsed >= 0 {
			offset = parsed
		}
	}
	return limit, offset
}

func extractItems(payload map[string]any) []map[string]any {
	if content, ok := payload["content"].([]any); ok {
		return normalizeMapSlice(content)
	}
	for _, key := range []string{"items", "users", "projects", "boards", "columns"} {
		if content, ok := payload[key].([]any); ok {
			return normalizeMapSlice(content)
		}
	}
	return nil
}

func normalizeMapSlice(items []any) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if typed, ok := item.(map[string]any); ok {
			result = append(result, typed)
		}
	}
	return result
}

func firstString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := stringFromAny(payload[key]); value != "" {
			return value
		}
	}
	return ""
}

func stringFromAny(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case fmt.Stringer:
		return strings.TrimSpace(typed.String())
	default:
		if typed != nil {
			return strings.TrimSpace(fmt.Sprint(typed))
		}
		return ""
	}
}

func normalizeOptionalString(values ...any) *string {
	for _, value := range values {
		if text := stringFromAny(value); text != "" {
			return &text
		}
	}
	return nil
}

func timeFromAny(value any) *time.Time {
	if text := stringFromAny(value); text != "" {
		for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"} {
			if parsed, err := time.Parse(layout, text); err == nil {
				return &parsed
			}
		}
	}
	return nil
}

func boolFromAny(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		typed = strings.ToLower(strings.TrimSpace(typed))
		return typed == "true" || typed == "1" || typed == "yes"
	case float64:
		return typed != 0
	default:
		return false
	}
}

func intFromAny(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		var parsed int
		if _, err := fmt.Sscanf(strings.TrimSpace(typed), "%d", &parsed); err == nil {
			return parsed
		}
	}
	return 0
}

func intPtrFromAny(value any) *int {
	switch typed := value.(type) {
	case int:
		return &typed
	case int32:
		v := int(typed)
		return &v
	case int64:
		v := int(typed)
		return &v
	case float64:
		v := int(typed)
		return &v
	case string:
		var v int
		if _, err := fmt.Sscanf(strings.TrimSpace(typed), "%d", &v); err == nil {
			return &v
		}
	}
	return nil
}

func defaultBaseURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "https://yougile.com"
	}
	if parsed, err := url.Parse(trimmed); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		return strings.TrimRight(trimmed, "/")
	}
	return "https://yougile.com"
}

func trimStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func stringPtr(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func lastN(value string, n int) string {
	value = strings.TrimSpace(value)
	if len(value) <= n {
		return value
	}
	return value[len(value)-n:]
}

func nullableJSON(payload json.RawMessage) string {
	if len(bytes.TrimSpace(payload)) == 0 {
		return "{}"
	}
	return string(payload)
}
