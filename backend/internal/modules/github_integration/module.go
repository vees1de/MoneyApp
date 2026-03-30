package github_integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
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
	ID                   uuid.UUID  `json:"id"`
	CompanyID            *uuid.UUID `json:"company_id,omitempty"`
	Title                string     `json:"title"`
	Provider             string     `json:"provider"`
	AuthMode             string     `json:"auth_mode"`
	BaseURL              string     `json:"base_url"`
	Status               string     `json:"status"`
	TokenLast4           *string    `json:"token_last4,omitempty"`
	GitHubAppID          *string    `json:"github_app_id,omitempty"`
	GitHubInstallationID *string    `json:"github_installation_id,omitempty"`
	CreatedBy            uuid.UUID  `json:"created_by"`
	LastSyncAt           *time.Time `json:"last_sync_at,omitempty"`
	LastSuccessSyncAt    *time.Time `json:"last_success_sync_at,omitempty"`
	LastError            *string    `json:"last_error,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type storedConnection struct {
	Connection
	TokenEncrypted *string
}

type GitHubUser struct {
	ID              uuid.UUID  `json:"id"`
	ConnectionID    uuid.UUID  `json:"connection_id"`
	GitHubUserID    int64      `json:"github_user_id"`
	Login           string     `json:"login"`
	Name            *string    `json:"name,omitempty"`
	Email           *string    `json:"email,omitempty"`
	AvatarURL       *string    `json:"avatar_url,omitempty"`
	HTMLURL         *string    `json:"html_url,omitempty"`
	Company         *string    `json:"company,omitempty"`
	Location        *string    `json:"location,omitempty"`
	Bio             *string    `json:"bio,omitempty"`
	Followers       *int       `json:"followers,omitempty"`
	Following       *int       `json:"following,omitempty"`
	PublicRepos     *int       `json:"public_repos,omitempty"`
	PublicGists     *int       `json:"public_gists,omitempty"`
	CreatedAtRemote *time.Time `json:"created_at_remote,omitempty"`
	UpdatedAtRemote *time.Time `json:"updated_at_remote,omitempty"`
	SyncedAt        time.Time  `json:"synced_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type Repository struct {
	ID               uuid.UUID  `json:"id"`
	ConnectionID     uuid.UUID  `json:"connection_id"`
	GitHubRepoID     int64      `json:"github_repo_id"`
	OwnerLogin       string     `json:"owner_login"`
	Name             string     `json:"name"`
	FullName         string     `json:"full_name"`
	Private          bool       `json:"private"`
	Archived         bool       `json:"archived"`
	Fork             bool       `json:"fork"`
	DefaultBranch    *string    `json:"default_branch,omitempty"`
	Language         *string    `json:"language,omitempty"`
	SizeKB           *int       `json:"size_kb,omitempty"`
	StargazersCount  *int       `json:"stargazers_count,omitempty"`
	WatchersCount    *int       `json:"watchers_count,omitempty"`
	ForksCount       *int       `json:"forks_count,omitempty"`
	OpenIssuesCount  *int       `json:"open_issues_count,omitempty"`
	SubscribersCount *int       `json:"subscribers_count,omitempty"`
	NetworkCount     *int       `json:"network_count,omitempty"`
	PushedAt         *time.Time `json:"pushed_at,omitempty"`
	CreatedAtRemote  *time.Time `json:"created_at_remote,omitempty"`
	UpdatedAtRemote  *time.Time `json:"updated_at_remote,omitempty"`
	HTMLURL          *string    `json:"html_url,omitempty"`
	SyncedAt         time.Time  `json:"synced_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type RepositoryLanguage struct {
	ID           uuid.UUID `json:"id"`
	RepositoryID uuid.UUID `json:"repository_id"`
	LanguageName string    `json:"language_name"`
	Bytes        int64     `json:"bytes"`
	Percent      string    `json:"percent"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RepositoryContributor struct {
	ID            uuid.UUID `json:"id"`
	RepositoryID  uuid.UUID `json:"repository_id"`
	GitHubUserID  *int64    `json:"github_user_id,omitempty"`
	GitHubLogin   string    `json:"github_login"`
	Contributions int       `json:"contributions"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type MappingView struct {
	ID             uuid.UUID `json:"id"`
	EmployeeUserID uuid.UUID `json:"employee_user_id"`
	EmployeeName   string    `json:"employee_name"`
	EmployeeEmail  string    `json:"employee_email"`
	GitHubLogin    string    `json:"github_login"`
	GitHubUserID   *int64    `json:"github_user_id,omitempty"`
	ProfileURL     *string   `json:"profile_url,omitempty"`
	MatchSource    string    `json:"match_source"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type SyncJob struct {
	ID           uuid.UUID       `json:"id"`
	ConnectionID uuid.UUID       `json:"connection_id"`
	JobType      string          `json:"job_type"`
	Status       string          `json:"status"`
	Cursor       json.RawMessage `json:"cursor,omitempty"`
	Progress     json.RawMessage `json:"progress"`
	Attempt      int             `json:"attempt"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	FinishedAt   *time.Time      `json:"finished_at,omitempty"`
	NextRetryAt  *time.Time      `json:"next_retry_at,omitempty"`
	ErrorText    *string         `json:"error_text,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type EmployeeLanguage struct {
	LanguageName string     `json:"name"`
	Bytes        int64      `json:"bytes"`
	Percent      string     `json:"percent"`
	ReposCount   int        `json:"reposCount"`
	LastUsedAt   *time.Time `json:"lastUsedAt,omitempty"`
}

type EmployeeStats struct {
	RepositoriesCount        int      `json:"repositoriesCount"`
	ActiveRepositoriesCount  int      `json:"activeRepositoriesCount"`
	Commits                  int      `json:"commits"`
	OpenedPRs                int      `json:"openedPRs"`
	MergedPRs                int      `json:"mergedPRs"`
	ReviewedPRs              int      `json:"reviewedPRs"`
	StarsReceived            int      `json:"starsReceived"`
	ForksReceived            int      `json:"forksReceived"`
	AvgRepoFreshnessDays     string   `json:"avgRepoFreshnessDays"`
	PrimaryLanguages         []string `json:"primaryLanguages"`
	EngineeringActivityScore string   `json:"engineeringActivityScore"`
	DataScope                string   `json:"dataScope"`
}

type EmployeeActivityPoint struct {
	Date        string `json:"date"`
	CommitCount int    `json:"commitCount"`
	OpenedPRs   int    `json:"openedPRs"`
	MergedPRs   int    `json:"mergedPRs"`
	ReviewedPRs int    `json:"reviewedPRs"`
}

type CreateConnectionRequest struct {
	Title    string  `json:"title" validate:"required"`
	AuthMode string  `json:"authMode" validate:"required,oneof=pat github_app oauth"`
	Token    *string `json:"token,omitempty"`
	BaseURL  string  `json:"baseUrl" validate:"omitempty,url"`
}

type TestConnectionRequest struct {
	Token   string `json:"token" validate:"required"`
	BaseURL string `json:"baseUrl" validate:"omitempty,url"`
}

type UpdateConnectionRequest struct {
	Title  *string `json:"title,omitempty"`
	Status *string `json:"status,omitempty" validate:"omitempty,oneof=active invalid revoked sync_error"`
}

type CreateMappingRequest struct {
	EmployeeUserID uuid.UUID `json:"employeeUserId" validate:"required"`
	GitHubLogin    string    `json:"githubLogin" validate:"required"`
}

type AutoMatchRequest struct {
	Strategy string `json:"strategy" validate:"required,oneof=email login domain"`
}

type SyncRequest struct {
	Mode                string `json:"mode" validate:"required,oneof=incremental full"`
	IncludeUsers        bool   `json:"includeUsers"`
	IncludeRepos        bool   `json:"includeRepos"`
	IncludeLanguages    bool   `json:"includeLanguages"`
	IncludeContributors bool   `json:"includeContributors"`
	IncludeActivity     bool   `json:"includeActivity"`
}

type TestConnectionResponse struct {
	OK        bool   `json:"ok"`
	Scopes    string `json:"scopes"`
	RateLimit string `json:"rateLimit"`
	Message   string `json:"message"`
}

type ImportResponse struct {
	Imported int `json:"imported"`
	Updated  int `json:"updated"`
	Failed   int `json:"failed"`
}

type RepositoryFilters struct {
	ConnectionID   *uuid.UUID
	EmployeeUserID *uuid.UUID
	Owner          string
	Language       string
	Archived       *bool
	Fork           *bool
	Visibility     string
	ActiveSince    *time.Time
	Limit          int
	Offset         int
}

type RepoSummary struct {
	RepositoriesCount int `json:"repositoriesCount"`
	PrivateCount      int `json:"privateCount"`
	PublicCount       int `json:"publicCount"`
}

type RepositoryOwnershipItem struct {
	EmployeeUserID     uuid.UUID `json:"employeeUserId"`
	EmployeeName       string    `json:"employeeName"`
	Repositories       int       `json:"repositories"`
	ActiveRepositories int       `json:"activeRepositories"`
}

type RepositoryHealthItem struct {
	RepositoryID  uuid.UUID `json:"repositoryId"`
	FullName      string    `json:"fullName"`
	Archived      bool      `json:"archived"`
	OpenIssues    int       `json:"openIssues"`
	FreshnessDays string    `json:"freshnessDays"`
}

type LanguageAnalyticsItem struct {
	Name    string `json:"name"`
	Percent string `json:"percent"`
}

type TeamAnalyticsEmployee struct {
	EmployeeUserID     uuid.UUID `json:"employeeUserId"`
	Name               string    `json:"name"`
	PrimaryLanguage    string    `json:"primaryLanguage"`
	ActiveRepositories int       `json:"activeRepositories"`
	Commits            int       `json:"commits"`
	MergedPRs          int       `json:"mergedPRs"`
}

type RepositoryRecord struct {
	Repository
}

type RepositoryDetail struct {
	Repository
	Languages    []RepositoryLanguage    `json:"languages"`
	Contributors []RepositoryContributor `json:"contributors"`
}

type RepositoryRef struct {
	ID         uuid.UUID
	OwnerLogin string
	Name       string
}

type GitHubClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewGitHubClient(baseURL, token string) *GitHubClient {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	return &GitHubClient{
		baseURL:    baseURL,
		token:      strings.TrimSpace(token),
		httpClient: &http.Client{Timeout: 25 * time.Second},
	}
}

func (c *GitHubClient) do(ctx context.Context, method, path string) ([]byte, http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.Header, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, resp.Header, httpx.NewError(resp.StatusCode, "github_request_failed", strings.TrimSpace(string(raw)))
	}
	return raw, resp.Header, nil
}

func (c *GitHubClient) GetCurrentUser(ctx context.Context) (map[string]any, http.Header, error) {
	raw, headers, err := c.do(ctx, http.MethodGet, "/user")
	if err != nil {
		return nil, headers, err
	}
	var payload map[string]any
	return payload, headers, json.Unmarshal(raw, &payload)
}

func (c *GitHubClient) GetUser(ctx context.Context, login string) (map[string]any, error) {
	raw, _, err := c.do(ctx, http.MethodGet, "/users/"+url.PathEscape(strings.TrimSpace(login)))
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	return payload, json.Unmarshal(raw, &payload)
}

func (c *GitHubClient) ListUserRepos(ctx context.Context, login string) ([]map[string]any, error) {
	path := "/users/" + url.PathEscape(strings.TrimSpace(login)) + "/repos?per_page=100&type=owner&sort=updated"
	if strings.TrimSpace(login) == "" {
		path = "/user/repos?per_page=100&type=all&sort=updated"
	}
	raw, _, err := c.do(ctx, http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	var items []map[string]any
	return items, json.Unmarshal(raw, &items)
}

func (c *GitHubClient) GetRepoLanguages(ctx context.Context, owner, name string) (map[string]int64, error) {
	raw, _, err := c.do(ctx, http.MethodGet, "/repos/"+url.PathEscape(owner)+"/"+url.PathEscape(name)+"/languages")
	if err != nil {
		return nil, err
	}
	var payload map[string]int64
	return payload, json.Unmarshal(raw, &payload)
}

func (c *GitHubClient) GetRepoContributors(ctx context.Context, owner, name string) ([]map[string]any, error) {
	raw, _, err := c.do(ctx, http.MethodGet, "/repos/"+url.PathEscape(owner)+"/"+url.PathEscape(name)+"/contributors?per_page=100")
	if err != nil {
		return nil, err
	}
	var items []map[string]any
	return items, json.Unmarshal(raw, &items)
}

type Repo struct {
	db *sql.DB
}

func NewRepository(database *sql.DB) *Repo {
	return &Repo{db: database}
}

func (r *Repo) base(exec ...db.DBTX) db.DBTX {
	if len(exec) > 0 && exec[0] != nil {
		return exec[0]
	}
	return r.db
}

func (r *Repo) CreateConnection(ctx context.Context, item storedConnection, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into github_connections (
			id, company_id, title, provider, auth_mode, base_url, status, token_encrypted, token_last4,
			github_app_id, github_installation_id, created_by, last_sync_at, last_success_sync_at, last_error, created_at, updated_at
		) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
	`, item.ID, item.CompanyID, item.Title, item.Provider, item.AuthMode, item.BaseURL, item.Status, item.TokenEncrypted,
		item.TokenLast4, item.GitHubAppID, item.GitHubInstallationID, item.CreatedBy, item.LastSyncAt, item.LastSuccessSyncAt, item.LastError, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repo) GetConnection(ctx context.Context, id uuid.UUID, exec ...db.DBTX) (storedConnection, error) {
	var item storedConnection
	err := r.base(exec...).QueryRowContext(ctx, `
			select id, company_id, title, provider, auth_mode, base_url, status, token_encrypted, token_last4,
			       github_app_id, github_installation_id, created_by, last_sync_at, last_success_sync_at, last_error, created_at, updated_at
			from github_connections where id = $1
		`, id).Scan(&item.ID, &item.CompanyID, &item.Title, &item.Provider, &item.AuthMode, &item.BaseURL, &item.Status,
		&item.TokenEncrypted, &item.TokenLast4, &item.GitHubAppID, &item.GitHubInstallationID, &item.CreatedBy,
		&item.LastSyncAt, &item.LastSuccessSyncAt, &item.LastError, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repo) ListConnections(ctx context.Context) ([]Connection, error) {
	rows, err := r.db.QueryContext(ctx, `
			select id, company_id, title, provider, auth_mode, base_url, status, token_last4,
			       github_app_id, github_installation_id, created_by, last_sync_at, last_success_sync_at, last_error, created_at, updated_at
			from github_connections order by created_at desc
		`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Connection
	for rows.Next() {
		var item Connection
		if err := rows.Scan(&item.ID, &item.CompanyID, &item.Title, &item.Provider, &item.AuthMode, &item.BaseURL, &item.Status,
			&item.TokenLast4, &item.GitHubAppID, &item.GitHubInstallationID, &item.CreatedBy, &item.LastSyncAt, &item.LastSuccessSyncAt, &item.LastError,
			&item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repo) UpdateConnection(ctx context.Context, item storedConnection, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update github_connections
		set title = $2, status = $3, base_url = $4, last_sync_at = $5, last_success_sync_at = $6, last_error = $7, updated_at = $8
		where id = $1
	`, item.ID, item.Title, item.Status, item.BaseURL, item.LastSyncAt, item.LastSuccessSyncAt, item.LastError, item.UpdatedAt)
	return err
}

func (r *Repo) RevokeConnection(ctx context.Context, id uuid.UUID, now time.Time) error {
	_, err := r.db.ExecContext(ctx, `update github_connections set status='revoked', updated_at=$2 where id=$1`, id, now)
	return err
}

func (r *Repo) UpsertGitHubUser(ctx context.Context, item GitHubUser, raw []byte, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into github_users (
			id, connection_id, github_user_id, login, name, email, avatar_url, html_url, company, location, bio,
			followers, following, public_repos, public_gists, created_at_remote, updated_at_remote, raw_payload, synced_at, created_at, updated_at
		) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18::jsonb,$19,$20,$21)
		on conflict (connection_id, github_user_id) do update
		set login = excluded.login,
		    name = excluded.name,
		    email = excluded.email,
		    avatar_url = excluded.avatar_url,
		    html_url = excluded.html_url,
		    company = excluded.company,
		    location = excluded.location,
		    bio = excluded.bio,
		    followers = excluded.followers,
		    following = excluded.following,
		    public_repos = excluded.public_repos,
		    public_gists = excluded.public_gists,
		    created_at_remote = excluded.created_at_remote,
		    updated_at_remote = excluded.updated_at_remote,
		    raw_payload = excluded.raw_payload,
		    synced_at = excluded.synced_at,
		    updated_at = excluded.updated_at
	`, item.ID, item.ConnectionID, item.GitHubUserID, item.Login, item.Name, item.Email, item.AvatarURL, item.HTMLURL, item.Company, item.Location, item.Bio,
		item.Followers, item.Following, item.PublicRepos, item.PublicGists, item.CreatedAtRemote, item.UpdatedAtRemote, raw, item.SyncedAt, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repo) ListGitHubUsers(ctx context.Context, connectionID *uuid.UUID, limit, offset int) ([]GitHubUser, error) {
	query := `
			select id, connection_id, github_user_id, login, name, email, avatar_url, html_url, company, location, bio,
			       followers, following, public_repos, public_gists, created_at_remote, updated_at_remote, synced_at, created_at, updated_at
			from github_users`
	args := []any{}
	if connectionID != nil {
		query += " where connection_id = $1"
		args = append(args, *connectionID)
	}
	args = append(args, limit, offset)
	query += fmt.Sprintf(" order by login asc limit $%d offset $%d", len(args)-1, len(args))
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GitHubUser
	for rows.Next() {
		var item GitHubUser
		if err := rows.Scan(&item.ID, &item.ConnectionID, &item.GitHubUserID, &item.Login, &item.Name, &item.Email, &item.AvatarURL, &item.HTMLURL, &item.Company, &item.Location, &item.Bio,
			&item.Followers, &item.Following, &item.PublicRepos, &item.PublicGists, &item.CreatedAtRemote, &item.UpdatedAtRemote, &item.SyncedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repo) FindGitHubUserByLogin(ctx context.Context, connectionID uuid.UUID, login string) (*GitHubUser, error) {
	var item GitHubUser
	err := r.db.QueryRowContext(ctx, `
			select id, connection_id, github_user_id, login, name, email, avatar_url, html_url, company, location, bio,
			       followers, following, public_repos, public_gists, created_at_remote, updated_at_remote, synced_at, created_at, updated_at
			from github_users
			where connection_id = $1 and lower(login) = lower($2)
		`, connectionID, login).Scan(&item.ID, &item.ConnectionID, &item.GitHubUserID, &item.Login, &item.Name, &item.Email, &item.AvatarURL, &item.HTMLURL, &item.Company, &item.Location, &item.Bio,
		&item.Followers, &item.Following, &item.PublicRepos, &item.PublicGists, &item.CreatedAtRemote, &item.UpdatedAtRemote, &item.SyncedAt, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *Repo) UpsertRepository(ctx context.Context, item Repository, raw []byte, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
			insert into github_repositories (
				id, connection_id, github_repo_id, owner_login, name, full_name, private, archived, fork, default_branch,
				language, size_kb, stargazers_count, watchers_count, forks_count, open_issues_count, subscribers_count,
				network_count, pushed_at, created_at_remote, updated_at_remote, html_url, raw_payload, synced_at, created_at, updated_at
			) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23::jsonb,$24,$25,$26)
			on conflict (connection_id, github_repo_id) do update
			set owner_login = excluded.owner_login,
			    name = excluded.name,
			    full_name = excluded.full_name,
			    private = excluded.private,
			    archived = excluded.archived,
			    fork = excluded.fork,
			    default_branch = excluded.default_branch,
			    language = excluded.language,
			    size_kb = excluded.size_kb,
			    stargazers_count = excluded.stargazers_count,
			    watchers_count = excluded.watchers_count,
			    forks_count = excluded.forks_count,
			    open_issues_count = excluded.open_issues_count,
			    subscribers_count = excluded.subscribers_count,
			    network_count = excluded.network_count,
			    pushed_at = excluded.pushed_at,
			    created_at_remote = excluded.created_at_remote,
			    updated_at_remote = excluded.updated_at_remote,
			    html_url = excluded.html_url,
			    raw_payload = excluded.raw_payload,
			    synced_at = excluded.synced_at,
			    updated_at = excluded.updated_at
		`, item.ID, item.ConnectionID, item.GitHubRepoID, item.OwnerLogin, item.Name, item.FullName, item.Private, item.Archived, item.Fork,
		item.DefaultBranch, item.Language, item.SizeKB, item.StargazersCount, item.WatchersCount, item.ForksCount, item.OpenIssuesCount, item.SubscribersCount,
		item.NetworkCount, item.PushedAt, item.CreatedAtRemote, item.UpdatedAtRemote, item.HTMLURL, raw, item.SyncedAt, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repo) ListRepositories(ctx context.Context, filters RepositoryFilters) ([]Repository, error) {
	query := `
			select gr.id, gr.connection_id, gr.github_repo_id, gr.owner_login, gr.name, gr.full_name, gr.private, gr.archived, gr.fork,
			       gr.default_branch, gr.language, gr.size_kb, gr.stargazers_count, gr.watchers_count, gr.forks_count, gr.open_issues_count,
			       gr.subscribers_count, gr.network_count, gr.pushed_at, gr.created_at_remote, gr.updated_at_remote, gr.html_url, gr.synced_at, gr.created_at, gr.updated_at
			from github_repositories gr`
	args := []any{}
	conds := []string{}
	if filters.EmployeeUserID != nil {
		query += `
				left join github_user_mappings gum on gum.connection_id = gr.connection_id and gum.is_active = true
				left join github_repository_contributors grc on grc.repository_id = gr.id`
		args = append(args, *filters.EmployeeUserID)
		conds = append(conds, fmt.Sprintf("(gum.employee_user_id = $%d and lower(gum.github_login) = lower(gr.owner_login) or grc.github_login in (select github_login from github_user_mappings where employee_user_id = $%d and is_active = true))", len(args), len(args)))
	}
	if filters.ConnectionID != nil {
		args = append(args, *filters.ConnectionID)
		conds = append(conds, fmt.Sprintf("gr.connection_id = $%d", len(args)))
	}
	if strings.TrimSpace(filters.Owner) != "" {
		args = append(args, strings.TrimSpace(filters.Owner))
		conds = append(conds, fmt.Sprintf("lower(gr.owner_login) = lower($%d)", len(args)))
	}
	if strings.TrimSpace(filters.Language) != "" {
		args = append(args, strings.TrimSpace(filters.Language))
		conds = append(conds, fmt.Sprintf("lower(coalesce(gr.language, '')) = lower($%d)", len(args)))
	}
	if filters.Archived != nil {
		args = append(args, *filters.Archived)
		conds = append(conds, fmt.Sprintf("gr.archived = $%d", len(args)))
	}
	if filters.Fork != nil {
		args = append(args, *filters.Fork)
		conds = append(conds, fmt.Sprintf("gr.fork = $%d", len(args)))
	}
	if strings.TrimSpace(filters.Visibility) == "public" {
		conds = append(conds, "gr.private = false")
	}
	if strings.TrimSpace(filters.Visibility) == "private" {
		conds = append(conds, "gr.private = true")
	}
	if filters.ActiveSince != nil {
		args = append(args, *filters.ActiveSince)
		conds = append(conds, fmt.Sprintf("gr.pushed_at >= $%d", len(args)))
	}
	if len(conds) > 0 {
		query += " where " + strings.Join(conds, " and ")
	}
	args = append(args, filters.Limit, filters.Offset)
	query += fmt.Sprintf(" order by gr.pushed_at desc nulls last, gr.full_name asc limit $%d offset $%d", len(args)-1, len(args))
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Repository
	for rows.Next() {
		var item Repository
		if err := rows.Scan(&item.ID, &item.ConnectionID, &item.GitHubRepoID, &item.OwnerLogin, &item.Name, &item.FullName, &item.Private, &item.Archived, &item.Fork,
			&item.DefaultBranch, &item.Language, &item.SizeKB, &item.StargazersCount, &item.WatchersCount, &item.ForksCount, &item.OpenIssuesCount,
			&item.SubscribersCount, &item.NetworkCount, &item.PushedAt, &item.CreatedAtRemote, &item.UpdatedAtRemote, &item.HTMLURL, &item.SyncedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repo) GetRepository(ctx context.Context, id uuid.UUID) (Repository, error) {
	var item Repository
	err := r.db.QueryRowContext(ctx, `
			select id, connection_id, github_repo_id, owner_login, name, full_name, private, archived, fork,
			       default_branch, language, size_kb, stargazers_count, watchers_count, forks_count, open_issues_count,
			       subscribers_count, network_count, pushed_at, created_at_remote, updated_at_remote, html_url, synced_at, created_at, updated_at
			from github_repositories where id = $1
		`, id).Scan(&item.ID, &item.ConnectionID, &item.GitHubRepoID, &item.OwnerLogin, &item.Name, &item.FullName, &item.Private, &item.Archived, &item.Fork,
		&item.DefaultBranch, &item.Language, &item.SizeKB, &item.StargazersCount, &item.WatchersCount, &item.ForksCount, &item.OpenIssuesCount,
		&item.SubscribersCount, &item.NetworkCount, &item.PushedAt, &item.CreatedAtRemote, &item.UpdatedAtRemote, &item.HTMLURL, &item.SyncedAt, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repo) ListRepositoryRefs(ctx context.Context, connectionID uuid.UUID) ([]RepositoryRef, error) {
	rows, err := r.db.QueryContext(ctx, `select id, owner_login, name from github_repositories where connection_id = $1`, connectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []RepositoryRef
	for rows.Next() {
		var item RepositoryRef
		if err := rows.Scan(&item.ID, &item.OwnerLogin, &item.Name); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repo) ReplaceRepositoryLanguages(ctx context.Context, repositoryID uuid.UUID, items []RepositoryLanguage, exec ...db.DBTX) error {
	if _, err := r.base(exec...).ExecContext(ctx, `delete from github_repository_languages where repository_id = $1`, repositoryID); err != nil {
		return err
	}
	for _, item := range items {
		if _, err := r.base(exec...).ExecContext(ctx, `
				insert into github_repository_languages (id, repository_id, language_name, bytes, percent, created_at, updated_at)
				values ($1,$2,$3,$4,$5,$6,$7)
			`, item.ID, item.RepositoryID, item.LanguageName, item.Bytes, item.Percent, item.CreatedAt, item.UpdatedAt); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repo) ListRepositoryLanguages(ctx context.Context, repositoryID uuid.UUID) ([]RepositoryLanguage, error) {
	rows, err := r.db.QueryContext(ctx, `
			select id, repository_id, language_name, bytes, percent::text, created_at, updated_at
			from github_repository_languages where repository_id = $1
			order by bytes desc, language_name asc
		`, repositoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []RepositoryLanguage
	for rows.Next() {
		var item RepositoryLanguage
		if err := rows.Scan(&item.ID, &item.RepositoryID, &item.LanguageName, &item.Bytes, &item.Percent, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repo) ReplaceRepositoryContributors(ctx context.Context, repositoryID uuid.UUID, items []RepositoryContributor, exec ...db.DBTX) error {
	if _, err := r.base(exec...).ExecContext(ctx, `delete from github_repository_contributors where repository_id = $1`, repositoryID); err != nil {
		return err
	}
	for _, item := range items {
		if _, err := r.base(exec...).ExecContext(ctx, `
				insert into github_repository_contributors (id, repository_id, github_user_id, github_login, contributions, created_at, updated_at)
				values ($1,$2,$3,$4,$5,$6,$7)
			`, item.ID, item.RepositoryID, item.GitHubUserID, item.GitHubLogin, item.Contributions, item.CreatedAt, item.UpdatedAt); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repo) ListRepositoryContributors(ctx context.Context, repositoryID uuid.UUID) ([]RepositoryContributor, error) {
	rows, err := r.db.QueryContext(ctx, `
			select id, repository_id, github_user_id, github_login, contributions, created_at, updated_at
			from github_repository_contributors where repository_id = $1
			order by contributions desc, github_login asc
		`, repositoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []RepositoryContributor
	for rows.Next() {
		var item RepositoryContributor
		if err := rows.Scan(&item.ID, &item.RepositoryID, &item.GitHubUserID, &item.GitHubLogin, &item.Contributions, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repo) CreateOrUpdateMapping(ctx context.Context, connectionID, employeeUserID uuid.UUID, login string, userID *int64, profileURL *string, source string, exec ...db.DBTX) error {
	now := time.Now().UTC()
	_, err := r.base(exec...).ExecContext(ctx, `
			insert into github_user_mappings (
				id, connection_id, employee_user_id, github_login, github_user_id, profile_url, match_source, is_active, created_at, updated_at
			)
			values ($1,$2,$3,$4,$5,$6,$7,true,$8,$8)
			on conflict (connection_id, employee_user_id) do update
			set github_login = excluded.github_login,
			    github_user_id = excluded.github_user_id,
			    profile_url = excluded.profile_url,
			    match_source = excluded.match_source,
			    is_active = true,
			    updated_at = excluded.updated_at
		`, uuid.New(), connectionID, employeeUserID, login, userID, profileURL, source, now)
	return err
}

func (r *Repo) DeleteMapping(ctx context.Context, connectionID, mappingID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `delete from github_user_mappings where connection_id = $1 and id = $2`, connectionID, mappingID)
	return err
}

func (r *Repo) ListMappings(ctx context.Context, connectionID uuid.UUID) ([]MappingView, error) {
	rows, err := r.db.QueryContext(ctx, `
			select gum.id, gum.employee_user_id, concat_ws(' ', ep.last_name, ep.first_name, ep.middle_name), u.email,
			       gum.github_login, gum.github_user_id, gum.profile_url, gum.match_source, gum.is_active, gum.created_at, gum.updated_at
			from github_user_mappings gum
			join users u on u.id = gum.employee_user_id
			left join employee_profiles ep on ep.user_id = gum.employee_user_id
			where gum.connection_id = $1
			order by 3 asc, 4 asc
		`, connectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []MappingView
	for rows.Next() {
		var item MappingView
		if err := rows.Scan(&item.ID, &item.EmployeeUserID, &item.EmployeeName, &item.EmployeeEmail, &item.GitHubLogin, &item.GitHubUserID, &item.ProfileURL, &item.MatchSource, &item.IsActive, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repo) ListInternalUsers(ctx context.Context) ([]struct {
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

func (r *Repo) ListGitHubUsersForMapping(ctx context.Context, connectionID uuid.UUID) ([]GitHubUser, error) {
	return r.ListGitHubUsers(ctx, &connectionID, 5000, 0)
}

func (r *Repo) UpsertLanguageProfile(ctx context.Context, connectionID, employeeUserID uuid.UUID, item EmployeeLanguage, exec ...db.DBTX) error {
	now := time.Now().UTC()
	_, err := r.base(exec...).ExecContext(ctx, `
			insert into github_employee_language_profiles (
				id, connection_id, employee_user_id, language_name, bytes, percent, repos_count, last_used_at, created_at, updated_at
			)
			values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$9)
			on conflict (connection_id, employee_user_id, language_name) do update
			set bytes = excluded.bytes,
			    percent = excluded.percent,
			    repos_count = excluded.repos_count,
			    last_used_at = excluded.last_used_at,
			    updated_at = excluded.updated_at
		`, uuid.New(), connectionID, employeeUserID, item.LanguageName, item.Bytes, item.Percent, item.ReposCount, item.LastUsedAt, now)
	return err
}

func (r *Repo) DeleteLanguageProfiles(ctx context.Context, connectionID, employeeUserID uuid.UUID, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `delete from github_employee_language_profiles where connection_id = $1 and employee_user_id = $2`, connectionID, employeeUserID)
	return err
}

func (r *Repo) ListEmployeeLanguages(ctx context.Context, connectionID, employeeUserID uuid.UUID) ([]EmployeeLanguage, error) {
	rows, err := r.db.QueryContext(ctx, `
			select language_name, bytes, percent::text, repos_count, last_used_at
			from github_employee_language_profiles
			where connection_id = $1 and employee_user_id = $2
			order by bytes desc, language_name asc
		`, connectionID, employeeUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []EmployeeLanguage
	for rows.Next() {
		var item EmployeeLanguage
		if err := rows.Scan(&item.LanguageName, &item.Bytes, &item.Percent, &item.ReposCount, &item.LastUsedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repo) ListActiveMappings(ctx context.Context, connectionID uuid.UUID) ([]MappingView, error) {
	return r.ListMappings(ctx, connectionID)
}

func (r *Repo) CreateSyncJob(ctx context.Context, item SyncJob) error {
	_, err := r.db.ExecContext(ctx, `
			insert into github_sync_jobs (
				id, connection_id, job_type, status, cursor, progress, attempt, started_at, finished_at, next_retry_at, error_text, created_at, updated_at
			)
			values ($1,$2,$3,$4,$5::jsonb,$6::jsonb,$7,$8,$9,$10,$11,$12,$13)
		`, item.ID, item.ConnectionID, item.JobType, item.Status, rawJSON(item.Cursor), rawJSON(item.Progress), item.Attempt, item.StartedAt, item.FinishedAt, item.NextRetryAt, item.ErrorText, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *Repo) GetSyncJob(ctx context.Context, id uuid.UUID) (SyncJob, error) {
	var item SyncJob
	err := r.db.QueryRowContext(ctx, `
			select id, connection_id, job_type, status, coalesce(cursor,'{}'::jsonb)::text, coalesce(progress,'{}'::jsonb)::text, attempt,
			       started_at, finished_at, next_retry_at, error_text, created_at, updated_at
			from github_sync_jobs where id = $1
		`, id).Scan(&item.ID, &item.ConnectionID, &item.JobType, &item.Status, &item.Cursor, &item.Progress, &item.Attempt, &item.StartedAt, &item.FinishedAt, &item.NextRetryAt, &item.ErrorText, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repo) UpdateSyncJob(ctx context.Context, id uuid.UUID, status string, progress []byte, startedAt, finishedAt *time.Time, errorText *string) error {
	_, err := r.db.ExecContext(ctx, `
			update github_sync_jobs
			set status = $2, progress = $3::jsonb, started_at = coalesce($4, started_at), finished_at = $5, error_text = $6, updated_at = now(),
			    attempt = case when $2 = 'processing' then attempt + 1 else attempt end
			where id = $1
		`, id, status, rawJSON(progress), startedAt, finishedAt, errorText)
	return err
}

func (r *Repo) ListCommitStats(ctx context.Context, connectionID, employeeUserID uuid.UUID, from, to time.Time) ([]EmployeeActivityPoint, error) {
	rows, err := r.db.QueryContext(ctx, `
			select to_char(metric_date, 'YYYY-MM-DD'), commit_count,
			       coalesce((select opened_prs from github_pull_request_stats prs where prs.connection_id = c.connection_id and prs.employee_user_id = c.employee_user_id and prs.metric_date = c.metric_date), 0),
			       coalesce((select merged_prs from github_pull_request_stats prs where prs.connection_id = c.connection_id and prs.employee_user_id = c.employee_user_id and prs.metric_date = c.metric_date), 0),
			       coalesce((select reviewed_prs from github_pull_request_stats prs where prs.connection_id = c.connection_id and prs.employee_user_id = c.employee_user_id and prs.metric_date = c.metric_date), 0)
			from github_commit_stats_daily c
			where c.connection_id = $1 and c.employee_user_id = $2 and c.metric_date between $3 and $4
			order by c.metric_date asc
		`, connectionID, employeeUserID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []EmployeeActivityPoint
	for rows.Next() {
		var item EmployeeActivityPoint
		if err := rows.Scan(&item.Date, &item.CommitCount, &item.OpenedPRs, &item.MergedPRs, &item.ReviewedPRs); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repo) EmployeeProfileData(ctx context.Context, connectionID, employeeUserID uuid.UUID) (map[string]any, error) {
	var profile = map[string]any{"employeeUserId": employeeUserID}
	row := r.db.QueryRowContext(ctx, `
			select gum.github_login, gu.name, gu.html_url, gu.avatar_url, gu.followers, gu.following, gu.public_repos
			from github_user_mappings gum
			left join github_users gu on gu.connection_id = gum.connection_id and lower(gu.login) = lower(gum.github_login)
			where gum.connection_id = $1 and gum.employee_user_id = $2 and gum.is_active = true
			limit 1
		`, connectionID, employeeUserID)
	var login string
	var name, profileURL, avatarURL *string
	var followers, following, publicRepos *int
	if err := row.Scan(&login, &name, &profileURL, &avatarURL, &followers, &following, &publicRepos); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			profile["github"] = nil
			return profile, nil
		}
		return nil, err
	}
	profile["github"] = map[string]any{
		"login":       login,
		"name":        name,
		"profileUrl":  profileURL,
		"avatarUrl":   avatarURL,
		"followers":   followers,
		"following":   following,
		"publicRepos": publicRepos,
	}
	return profile, nil
}

type Service struct {
	db    *sql.DB
	repo  *Repo
	queue *worker.Queue
	clock clock.Clock
}

func NewService(database *sql.DB, repo *Repo, queue *worker.Queue, appClock clock.Clock) *Service {
	return &Service{db: database, repo: repo, queue: queue, clock: appClock}
}

func (s *Service) CreateConnection(ctx context.Context, principal platformauth.Principal, req CreateConnectionRequest) (Connection, error) {
	if req.AuthMode == "pat" && (req.Token == nil || strings.TrimSpace(*req.Token) == "") {
		return Connection{}, httpx.BadRequest("missing_token", "token is required for pat auth mode")
	}
	baseURL := defaultGitHubBaseURL(req.BaseURL)
	token := ""
	if req.Token != nil {
		token = strings.TrimSpace(*req.Token)
	}
	client := NewGitHubClient(baseURL, token)
	if _, _, err := client.GetCurrentUser(ctx); err != nil {
		return Connection{}, err
	}
	now := s.clock.Now()
	item := storedConnection{
		Connection: Connection{
			ID:         uuid.New(),
			Title:      strings.TrimSpace(req.Title),
			Provider:   "github",
			AuthMode:   req.AuthMode,
			BaseURL:    baseURL,
			Status:     "active",
			TokenLast4: stringPtr(lastN(token, 4)),
			CreatedBy:  principal.UserID,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		TokenEncrypted: stringPtr(token),
	}
	if err := s.repo.CreateConnection(ctx, item); err != nil {
		return Connection{}, err
	}
	return item.Connection, nil
}

func (s *Service) TestConnection(ctx context.Context, req TestConnectionRequest) (TestConnectionResponse, error) {
	_, headers, err := NewGitHubClient(defaultGitHubBaseURL(req.BaseURL), req.Token).GetCurrentUser(ctx)
	if err != nil {
		return TestConnectionResponse{}, err
	}
	return TestConnectionResponse{
		OK:        true,
		Scopes:    headers.Get("X-OAuth-Scopes"),
		RateLimit: headers.Get("X-RateLimit-Limit"),
		Message:   "Connection verified",
	}, nil
}

func (s *Service) ListConnections(ctx context.Context) ([]Connection, error) {
	return s.repo.ListConnections(ctx)
}
func (s *Service) GetConnection(ctx context.Context, id uuid.UUID) (Connection, error) {
	item, err := s.repo.GetConnection(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Connection{}, httpx.NotFound("github_connection_not_found", "github connection not found")
		}
		return Connection{}, err
	}
	return item.Connection, nil
}

func (s *Service) UpdateConnection(ctx context.Context, id uuid.UUID, req UpdateConnectionRequest) (Connection, error) {
	item, err := s.repo.GetConnection(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Connection{}, httpx.NotFound("github_connection_not_found", "github connection not found")
		}
		return Connection{}, err
	}
	if req.Title != nil {
		item.Title = strings.TrimSpace(*req.Title)
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

func (s *Service) DeleteConnection(ctx context.Context, id uuid.UUID) error {
	return s.repo.RevokeConnection(ctx, id, s.clock.Now())
}

func (s *Service) ImportUsers(ctx context.Context, connectionID uuid.UUID) (ImportResponse, error) {
	conn, err := s.repo.GetConnection(ctx, connectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ImportResponse{}, httpx.NotFound("github_connection_not_found", "github connection not found")
		}
		return ImportResponse{}, err
	}
	client := NewGitHubClient(conn.BaseURL, derefString(conn.TokenEncrypted))
	now := s.clock.Now()
	result := ImportResponse{}
	current, _, err := client.GetCurrentUser(ctx)
	if err == nil {
		if err := s.upsertGitHubUserPayload(ctx, connectionID, current, now); err == nil {
			result.Imported++
		}
	}
	mappings, err := s.repo.ListMappings(ctx, connectionID)
	if err != nil {
		return result, err
	}
	for _, mapping := range mappings {
		payload, err := client.GetUser(ctx, mapping.GitHubLogin)
		if err != nil {
			result.Failed++
			continue
		}
		if err := s.upsertGitHubUserPayload(ctx, connectionID, payload, now); err != nil {
			result.Failed++
			continue
		}
		result.Updated++
	}
	return result, nil
}

func (s *Service) upsertGitHubUserPayload(ctx context.Context, connectionID uuid.UUID, payload map[string]any, now time.Time) error {
	item := GitHubUser{
		ID:              uuid.New(),
		ConnectionID:    connectionID,
		GitHubUserID:    int64FromAny(payload["id"]),
		Login:           stringFromAny(payload["login"]),
		Name:            optionalString(payload["name"]),
		Email:           optionalString(payload["email"]),
		AvatarURL:       optionalString(payload["avatar_url"]),
		HTMLURL:         optionalString(payload["html_url"]),
		Company:         optionalString(payload["company"]),
		Location:        optionalString(payload["location"]),
		Bio:             optionalString(payload["bio"]),
		Followers:       intPtrFromAny(payload["followers"]),
		Following:       intPtrFromAny(payload["following"]),
		PublicRepos:     intPtrFromAny(payload["public_repos"]),
		PublicGists:     intPtrFromAny(payload["public_gists"]),
		CreatedAtRemote: timePtrFromAny(payload["created_at"]),
		UpdatedAtRemote: timePtrFromAny(payload["updated_at"]),
		SyncedAt:        now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if item.GitHubUserID == 0 || item.Login == "" {
		return httpx.BadRequest("github_user_invalid", "github user payload is missing id or login")
	}
	raw, _ := json.Marshal(payload)
	return s.repo.UpsertGitHubUser(ctx, item, raw)
}

func (s *Service) ImportRepos(ctx context.Context, connectionID uuid.UUID) (ImportResponse, error) {
	conn, err := s.repo.GetConnection(ctx, connectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ImportResponse{}, httpx.NotFound("github_connection_not_found", "github connection not found")
		}
		return ImportResponse{}, err
	}
	client := NewGitHubClient(conn.BaseURL, derefString(conn.TokenEncrypted))
	now := s.clock.Now()
	result := ImportResponse{}
	mappings, _ := s.repo.ListMappings(ctx, connectionID)
	importedAny := false
	for _, mapping := range mappings {
		repos, err := client.ListUserRepos(ctx, mapping.GitHubLogin)
		if err != nil {
			result.Failed++
			continue
		}
		for _, payload := range repos {
			if err := s.upsertRepositoryPayload(ctx, connectionID, payload, now); err != nil {
				result.Failed++
				continue
			}
			importedAny = true
			result.Imported++
		}
	}
	if !importedAny {
		repos, err := client.ListUserRepos(ctx, "")
		if err != nil {
			return result, err
		}
		for _, payload := range repos {
			if err := s.upsertRepositoryPayload(ctx, connectionID, payload, now); err != nil {
				result.Failed++
				continue
			}
			result.Imported++
		}
	}
	return result, nil
}

func (s *Service) upsertRepositoryPayload(ctx context.Context, connectionID uuid.UUID, payload map[string]any, now time.Time) error {
	ownerLogin := ""
	if owner, ok := payload["owner"].(map[string]any); ok {
		ownerLogin = stringFromAny(owner["login"])
	}
	item := Repository{
		ID:               uuid.New(),
		ConnectionID:     connectionID,
		GitHubRepoID:     int64FromAny(payload["id"]),
		OwnerLogin:       ownerLogin,
		Name:             stringFromAny(payload["name"]),
		FullName:         stringFromAny(payload["full_name"]),
		Private:          boolFromAny(payload["private"]),
		Archived:         boolFromAny(payload["archived"]),
		Fork:             boolFromAny(payload["fork"]),
		DefaultBranch:    optionalString(payload["default_branch"]),
		Language:         optionalString(payload["language"]),
		SizeKB:           intPtrFromAny(payload["size"]),
		StargazersCount:  intPtrFromAny(payload["stargazers_count"]),
		WatchersCount:    intPtrFromAny(payload["watchers_count"]),
		ForksCount:       intPtrFromAny(payload["forks_count"]),
		OpenIssuesCount:  intPtrFromAny(payload["open_issues_count"]),
		SubscribersCount: intPtrFromAny(payload["subscribers_count"]),
		NetworkCount:     intPtrFromAny(payload["network_count"]),
		PushedAt:         timePtrFromAny(payload["pushed_at"]),
		CreatedAtRemote:  timePtrFromAny(payload["created_at"]),
		UpdatedAtRemote:  timePtrFromAny(payload["updated_at"]),
		HTMLURL:          optionalString(payload["html_url"]),
		SyncedAt:         now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if item.GitHubRepoID == 0 || item.FullName == "" || item.OwnerLogin == "" {
		return httpx.BadRequest("github_repo_invalid", "github repo payload is missing id, full_name, or owner")
	}
	raw, _ := json.Marshal(payload)
	return s.repo.UpsertRepository(ctx, item, raw)
}

func (s *Service) ImportLanguages(ctx context.Context, connectionID uuid.UUID) (ImportResponse, error) {
	conn, err := s.repo.GetConnection(ctx, connectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ImportResponse{}, httpx.NotFound("github_connection_not_found", "github connection not found")
		}
		return ImportResponse{}, err
	}
	client := NewGitHubClient(conn.BaseURL, derefString(conn.TokenEncrypted))
	repos, err := s.repo.ListRepositoryRefs(ctx, connectionID)
	if err != nil {
		return ImportResponse{}, err
	}
	now := s.clock.Now()
	result := ImportResponse{}
	for _, repo := range repos {
		payload, err := client.GetRepoLanguages(ctx, repo.OwnerLogin, repo.Name)
		if err != nil {
			result.Failed++
			continue
		}
		total := int64(0)
		for _, bytes := range payload {
			total += bytes
		}
		items := make([]RepositoryLanguage, 0, len(payload))
		for language, bytes := range payload {
			percent := formatDecimal(percentOf(bytes, total))
			items = append(items, RepositoryLanguage{
				ID:           uuid.New(),
				RepositoryID: repo.ID,
				LanguageName: language,
				Bytes:        bytes,
				Percent:      percent,
				CreatedAt:    now,
				UpdatedAt:    now,
			})
		}
		if err := s.repo.ReplaceRepositoryLanguages(ctx, repo.ID, items); err != nil {
			result.Failed++
			continue
		}
		result.Imported++
	}
	_ = s.rebuildLanguageProfiles(ctx, connectionID)
	return result, nil
}

func (s *Service) ImportContributors(ctx context.Context, connectionID uuid.UUID) (ImportResponse, error) {
	conn, err := s.repo.GetConnection(ctx, connectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ImportResponse{}, httpx.NotFound("github_connection_not_found", "github connection not found")
		}
		return ImportResponse{}, err
	}
	client := NewGitHubClient(conn.BaseURL, derefString(conn.TokenEncrypted))
	repos, err := s.repo.ListRepositoryRefs(ctx, connectionID)
	if err != nil {
		return ImportResponse{}, err
	}
	now := s.clock.Now()
	result := ImportResponse{}
	for _, repo := range repos {
		payload, err := client.GetRepoContributors(ctx, repo.OwnerLogin, repo.Name)
		if err != nil {
			result.Failed++
			continue
		}
		items := make([]RepositoryContributor, 0, len(payload))
		for _, raw := range payload {
			userID := int64FromAny(raw["id"])
			login := stringFromAny(raw["login"])
			if login == "" {
				continue
			}
			var ref *int64
			if userID != 0 {
				ref = &userID
			}
			items = append(items, RepositoryContributor{
				ID:            uuid.New(),
				RepositoryID:  repo.ID,
				GitHubUserID:  ref,
				GitHubLogin:   login,
				Contributions: intFromAny(raw["contributions"]),
				CreatedAt:     now,
				UpdatedAt:     now,
			})
		}
		if err := s.repo.ReplaceRepositoryContributors(ctx, repo.ID, items); err != nil {
			result.Failed++
			continue
		}
		result.Imported++
	}
	_ = s.rebuildLanguageProfiles(ctx, connectionID)
	return result, nil
}

func (s *Service) rebuildLanguageProfiles(ctx context.Context, connectionID uuid.UUID) error {
	mappings, err := s.repo.ListMappings(ctx, connectionID)
	if err != nil {
		return err
	}
	allRepos, err := s.repo.ListRepositories(ctx, RepositoryFilters{ConnectionID: &connectionID, Limit: 10000, Offset: 0})
	if err != nil {
		return err
	}
	for _, mapping := range mappings {
		if err := s.repo.DeleteLanguageProfiles(ctx, connectionID, mapping.EmployeeUserID); err != nil {
			return err
		}
		totals := map[string]int64{}
		repoCountByLanguage := map[string]int{}
		var lastUsed map[string]*time.Time = map[string]*time.Time{}
		for _, repo := range allRepos {
			inScope := strings.EqualFold(repo.OwnerLogin, mapping.GitHubLogin)
			if !inScope {
				contributors, err := s.repo.ListRepositoryContributors(ctx, repo.ID)
				if err == nil {
					for _, contributor := range contributors {
						if strings.EqualFold(contributor.GitHubLogin, mapping.GitHubLogin) {
							inScope = true
							break
						}
					}
				}
			}
			if !inScope {
				continue
			}
			languages, err := s.repo.ListRepositoryLanguages(ctx, repo.ID)
			if err != nil {
				continue
			}
			for _, language := range languages {
				totals[language.LanguageName] += language.Bytes
				repoCountByLanguage[language.LanguageName]++
				if repo.PushedAt != nil {
					last := lastUsed[language.LanguageName]
					if last == nil || repo.PushedAt.After(*last) {
						copied := *repo.PushedAt
						lastUsed[language.LanguageName] = &copied
					}
				}
			}
		}
		totalBytes := int64(0)
		for _, bytes := range totals {
			totalBytes += bytes
		}
		for name, bytes := range totals {
			item := EmployeeLanguage{
				LanguageName: name,
				Bytes:        bytes,
				Percent:      formatDecimal(percentOf(bytes, totalBytes)),
				ReposCount:   repoCountByLanguage[name],
				LastUsedAt:   lastUsed[name],
			}
			if err := s.repo.UpsertLanguageProfile(ctx, connectionID, mapping.EmployeeUserID, item); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) StartSync(ctx context.Context, connectionID uuid.UUID, req SyncRequest) (SyncJob, error) {
	now := s.clock.Now()
	jobType := "full_sync"
	if req.IncludeUsers && !req.IncludeRepos && !req.IncludeLanguages && !req.IncludeActivity && !req.IncludeContributors {
		jobType = "users_sync"
	} else if !req.IncludeUsers && req.IncludeRepos && !req.IncludeLanguages && !req.IncludeContributors && !req.IncludeActivity {
		jobType = "repos_sync"
	} else if !req.IncludeUsers && !req.IncludeRepos && req.IncludeLanguages && !req.IncludeContributors && !req.IncludeActivity {
		jobType = "languages_sync"
	} else if req.IncludeActivity && !req.IncludeUsers && !req.IncludeRepos && !req.IncludeLanguages {
		jobType = "activity_sync"
	}
	progress, _ := json.Marshal(map[string]any{
		"mode": req.Mode,
	})
	job := SyncJob{
		ID:           uuid.New(),
		ConnectionID: connectionID,
		JobType:      jobType,
		Status:       "pending",
		Cursor:       json.RawMessage(`{}`),
		Progress:     progress,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.CreateSyncJob(ctx, job); err != nil {
		return SyncJob{}, err
	}
	idKey := "github-sync:" + job.ID.String()
	if err := s.queue.Enqueue(ctx, s.db, "integrations", "github_sync", map[string]any{
		"sync_job_id":          job.ID,
		"connection_id":        connectionID,
		"include_users":        req.IncludeUsers,
		"include_repos":        req.IncludeRepos,
		"include_languages":    req.IncludeLanguages,
		"include_contributors": req.IncludeContributors,
		"include_activity":     req.IncludeActivity,
	}, &idKey, now); err != nil {
		return SyncJob{}, err
	}
	return job, nil
}

func (s *Service) GetSyncJob(ctx context.Context, id uuid.UUID) (SyncJob, error) {
	job, err := s.repo.GetSyncJob(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SyncJob{}, httpx.NotFound("github_sync_job_not_found", "github sync job not found")
		}
		return SyncJob{}, err
	}
	return job, nil
}

func (s *Service) ProcessSyncJob(ctx context.Context, job worker.Job) error {
	var payload struct {
		SyncJobID           uuid.UUID `json:"sync_job_id"`
		ConnectionID        uuid.UUID `json:"connection_id"`
		IncludeUsers        bool      `json:"include_users"`
		IncludeRepos        bool      `json:"include_repos"`
		IncludeLanguages    bool      `json:"include_languages"`
		IncludeContributors bool      `json:"include_contributors"`
		IncludeActivity     bool      `json:"include_activity"`
	}
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return err
	}
	now := s.clock.Now()
	progress := map[string]any{"status": "processing"}
	progressJSON, _ := json.Marshal(progress)
	if err := s.repo.UpdateSyncJob(ctx, payload.SyncJobID, "processing", progressJSON, &now, nil, nil); err != nil {
		return err
	}
	if payload.IncludeUsers {
		res, err := s.ImportUsers(ctx, payload.ConnectionID)
		if err != nil {
			msg := err.Error()
			_ = s.repo.UpdateSyncJob(ctx, payload.SyncJobID, "failed", progressJSON, nil, &now, &msg)
			return err
		}
		progress["users"] = res
	}
	if payload.IncludeRepos {
		res, err := s.ImportRepos(ctx, payload.ConnectionID)
		if err != nil {
			msg := err.Error()
			_ = s.repo.UpdateSyncJob(ctx, payload.SyncJobID, "failed", progressJSON, nil, &now, &msg)
			return err
		}
		progress["repos"] = res
	}
	if payload.IncludeContributors {
		res, err := s.ImportContributors(ctx, payload.ConnectionID)
		if err != nil {
			msg := err.Error()
			_ = s.repo.UpdateSyncJob(ctx, payload.SyncJobID, "failed", progressJSON, nil, &now, &msg)
			return err
		}
		progress["contributors"] = res
	}
	if payload.IncludeLanguages {
		res, err := s.ImportLanguages(ctx, payload.ConnectionID)
		if err != nil {
			msg := err.Error()
			_ = s.repo.UpdateSyncJob(ctx, payload.SyncJobID, "failed", progressJSON, nil, &now, &msg)
			return err
		}
		progress["languages"] = res
	}
	if payload.IncludeActivity {
		progress["activity"] = "storage-ready"
	}
	progress["status"] = "done"
	progressJSON, _ = json.Marshal(progress)
	return s.repo.UpdateSyncJob(ctx, payload.SyncJobID, "done", progressJSON, nil, &now, nil)
}

func (s *Service) ListMappings(ctx context.Context, connectionID uuid.UUID) ([]MappingView, error) {
	return s.repo.ListMappings(ctx, connectionID)
}

func (s *Service) CreateMapping(ctx context.Context, connectionID uuid.UUID, req CreateMappingRequest) error {
	var githubUserID *int64
	var profileURL *string
	if item, err := s.repo.FindGitHubUserByLogin(ctx, connectionID, req.GitHubLogin); err == nil && item != nil {
		githubUserID = &item.GitHubUserID
		profileURL = item.HTMLURL
	}
	return s.repo.CreateOrUpdateMapping(ctx, connectionID, req.EmployeeUserID, strings.TrimSpace(req.GitHubLogin), githubUserID, profileURL, "manual")
}

func (s *Service) DeleteMapping(ctx context.Context, connectionID, mappingID uuid.UUID) error {
	return s.repo.DeleteMapping(ctx, connectionID, mappingID)
}

func (s *Service) AutoMatch(ctx context.Context, connectionID uuid.UUID, strategy string) (map[string]any, error) {
	internalUsers, err := s.repo.ListInternalUsers(ctx)
	if err != nil {
		return nil, err
	}
	githubUsers, err := s.repo.ListGitHubUsersForMapping(ctx, connectionID)
	if err != nil {
		return nil, err
	}
	matched := 0
	internalMatched := map[uuid.UUID]struct{}{}
	githubMatched := map[string]struct{}{}
	for _, internal := range internalUsers {
		localPart, domain := splitEmail(internal.Email)
		for _, gh := range githubUsers {
			if shouldMatch(strategy, internal.Email, localPart, domain, gh) {
				_ = s.repo.CreateOrUpdateMapping(ctx, connectionID, internal.UserID, gh.Login, &gh.GitHubUserID, gh.HTMLURL, strategy)
				matched++
				internalMatched[internal.UserID] = struct{}{}
				githubMatched[strings.ToLower(gh.Login)] = struct{}{}
				break
			}
		}
	}
	return map[string]any{
		"matched":           matched,
		"unmatchedInternal": max(0, len(internalUsers)-len(internalMatched)),
		"unmatchedGithub":   max(0, len(githubUsers)-len(githubMatched)),
	}, nil
}

func (s *Service) ListGitHubUsers(ctx context.Context, connectionID *uuid.UUID, limit, offset int) ([]GitHubUser, error) {
	return s.repo.ListGitHubUsers(ctx, connectionID, limit, offset)
}

func (s *Service) ListRepositories(ctx context.Context, filters RepositoryFilters) ([]Repository, error) {
	return s.repo.ListRepositories(ctx, filters)
}

func (s *Service) GetRepositoryDetail(ctx context.Context, repoID uuid.UUID) (RepositoryDetail, error) {
	repo, err := s.repo.GetRepository(ctx, repoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RepositoryDetail{}, httpx.NotFound("github_repo_not_found", "github repository not found")
		}
		return RepositoryDetail{}, err
	}
	languages, _ := s.repo.ListRepositoryLanguages(ctx, repoID)
	contributors, _ := s.repo.ListRepositoryContributors(ctx, repoID)
	return RepositoryDetail{Repository: repo, Languages: languages, Contributors: contributors}, nil
}

func (s *Service) GetRepositoryLanguages(ctx context.Context, repoID uuid.UUID) ([]RepositoryLanguage, error) {
	return s.repo.ListRepositoryLanguages(ctx, repoID)
}

func (s *Service) GetRepositoryContributors(ctx context.Context, repoID uuid.UUID) ([]RepositoryContributor, error) {
	return s.repo.ListRepositoryContributors(ctx, repoID)
}

func (s *Service) GetEmployeeProfile(ctx context.Context, connectionID, employeeUserID uuid.UUID) (map[string]any, error) {
	return s.repo.EmployeeProfileData(ctx, connectionID, employeeUserID)
}

func (s *Service) GetEmployeeLanguages(ctx context.Context, connectionID, employeeUserID uuid.UUID) (map[string]any, error) {
	languages, err := s.repo.ListEmployeeLanguages(ctx, connectionID, employeeUserID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"employeeUserId": employeeUserID,
		"languages":      languages,
	}, nil
}

func (s *Service) GetEmployeeStats(ctx context.Context, connectionID, employeeUserID uuid.UUID, from, to time.Time) (map[string]any, error) {
	repos, err := s.repo.ListRepositories(ctx, RepositoryFilters{ConnectionID: &connectionID, EmployeeUserID: &employeeUserID, Limit: 10000, Offset: 0})
	if err != nil {
		return nil, err
	}
	languages, _ := s.repo.ListEmployeeLanguages(ctx, connectionID, employeeUserID)
	activity, _ := s.repo.ListCommitStats(ctx, connectionID, employeeUserID, from, to)
	stats := EmployeeStats{DataScope: "public_only"}
	privateSeen := false
	freshnessSum := 0.0
	freshnessCount := 0
	activeRepos := 0
	for _, repo := range repos {
		stats.RepositoriesCount++
		if repo.Private {
			privateSeen = true
		}
		if repo.PushedAt != nil {
			days := time.Since(*repo.PushedAt).Hours() / 24
			freshnessSum += days
			freshnessCount++
			if repo.PushedAt.After(from) {
				activeRepos++
			}
		}
		stats.StarsReceived += intOrZero(repo.StargazersCount)
		stats.ForksReceived += intOrZero(repo.ForksCount)
	}
	stats.ActiveRepositoriesCount = activeRepos
	if freshnessCount > 0 {
		stats.AvgRepoFreshnessDays = formatDecimal(freshnessSum / float64(freshnessCount))
	} else {
		stats.AvgRepoFreshnessDays = "0"
	}
	for _, point := range activity {
		stats.Commits += point.CommitCount
		stats.OpenedPRs += point.OpenedPRs
		stats.MergedPRs += point.MergedPRs
		stats.ReviewedPRs += point.ReviewedPRs
	}
	for idx, language := range languages {
		if idx >= 3 {
			break
		}
		stats.PrimaryLanguages = append(stats.PrimaryLanguages, language.LanguageName)
	}
	if privateSeen {
		stats.DataScope = "private_enabled"
	}
	stats.EngineeringActivityScore = formatDecimal(engineeringScore(stats))
	return map[string]any{
		"employeeUserId": employeeUserID,
		"period": map[string]any{
			"from": from.Format("2006-01-02"),
			"to":   to.Format("2006-01-02"),
		},
		"stats": stats,
	}, nil
}

func (s *Service) GetEmployeeActivity(ctx context.Context, connectionID, employeeUserID uuid.UUID, from, to time.Time) (map[string]any, error) {
	items, err := s.repo.ListCommitStats(ctx, connectionID, employeeUserID, from, to)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"employeeUserId": employeeUserID,
		"period": map[string]any{
			"from": from.Format("2006-01-02"),
			"to":   to.Format("2006-01-02"),
		},
		"items": items,
	}, nil
}

func (s *Service) TeamAnalytics(ctx context.Context, connectionID uuid.UUID, from, to time.Time) (map[string]any, error) {
	mappings, err := s.repo.ListMappings(ctx, connectionID)
	if err != nil {
		return nil, err
	}
	summary := map[string]int{
		"employees":          len(mappings),
		"activeRepositories": 0,
		"commitsThisMonth":   0,
		"mergedPRsThisMonth": 0,
	}
	employees := make([]TeamAnalyticsEmployee, 0, len(mappings))
	for _, mapping := range mappings {
		statsPayload, err := s.GetEmployeeStats(ctx, connectionID, mapping.EmployeeUserID, from, to)
		if err != nil {
			continue
		}
		statsMap := statsPayload["stats"].(EmployeeStats)
		summary["activeRepositories"] += statsMap.ActiveRepositoriesCount
		summary["commitsThisMonth"] += statsMap.Commits
		summary["mergedPRsThisMonth"] += statsMap.MergedPRs
		primary := ""
		if len(statsMap.PrimaryLanguages) > 0 {
			primary = statsMap.PrimaryLanguages[0]
		}
		employees = append(employees, TeamAnalyticsEmployee{
			EmployeeUserID:     mapping.EmployeeUserID,
			Name:               mapping.EmployeeName,
			PrimaryLanguage:    primary,
			ActiveRepositories: statsMap.ActiveRepositoriesCount,
			Commits:            statsMap.Commits,
			MergedPRs:          statsMap.MergedPRs,
		})
	}
	return map[string]any{"summary": summary, "employees": employees}, nil
}

func (s *Service) LanguageAnalytics(ctx context.Context, connectionID uuid.UUID) (map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `
		select language_name, coalesce(sum(bytes),0) as total_bytes
		from github_employee_language_profiles
		where connection_id = $1
		group by language_name
		order by total_bytes desc
	`, connectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type row struct {
		name  string
		bytes int64
	}
	var rowsData []row
	total := int64(0)
	for rows.Next() {
		var item row
		if err := rows.Scan(&item.name, &item.bytes); err != nil {
			return nil, err
		}
		rowsData = append(rowsData, item)
		total += item.bytes
	}
	items := make([]LanguageAnalyticsItem, 0, len(rowsData))
	for _, item := range rowsData {
		items = append(items, LanguageAnalyticsItem{Name: item.name, Percent: formatDecimal(percentOf(item.bytes, total))})
	}
	return map[string]any{"languages": items}, nil
}

func (s *Service) TopLanguages(ctx context.Context, connectionID uuid.UUID) (map[string]any, error) {
	return s.LanguageAnalytics(ctx, connectionID)
}

func (s *Service) RepositoryHealth(ctx context.Context, connectionID uuid.UUID) ([]RepositoryHealthItem, error) {
	repos, err := s.repo.ListRepositories(ctx, RepositoryFilters{ConnectionID: &connectionID, Limit: 10000, Offset: 0})
	if err != nil {
		return nil, err
	}
	items := make([]RepositoryHealthItem, 0, len(repos))
	for _, repo := range repos {
		freshness := "0"
		if repo.PushedAt != nil {
			freshness = formatDecimal(time.Since(*repo.PushedAt).Hours() / 24)
		}
		items = append(items, RepositoryHealthItem{
			RepositoryID:  repo.ID,
			FullName:      repo.FullName,
			Archived:      repo.Archived,
			OpenIssues:    intOrZero(repo.OpenIssuesCount),
			FreshnessDays: freshness,
		})
	}
	return items, nil
}

func (s *Service) RepositoryOwnership(ctx context.Context, connectionID uuid.UUID, from time.Time) ([]RepositoryOwnershipItem, error) {
	mappings, err := s.repo.ListMappings(ctx, connectionID)
	if err != nil {
		return nil, err
	}
	items := make([]RepositoryOwnershipItem, 0, len(mappings))
	for _, mapping := range mappings {
		repos, err := s.repo.ListRepositories(ctx, RepositoryFilters{ConnectionID: &connectionID, EmployeeUserID: &mapping.EmployeeUserID, Limit: 10000, Offset: 0})
		if err != nil {
			continue
		}
		active := 0
		for _, repo := range repos {
			if repo.PushedAt != nil && repo.PushedAt.After(from) {
				active++
			}
		}
		items = append(items, RepositoryOwnershipItem{
			EmployeeUserID:     mapping.EmployeeUserID,
			EmployeeName:       mapping.EmployeeName,
			Repositories:       len(repos),
			ActiveRepositories: active,
		})
	}
	return items, nil
}

type Handler struct {
	service   *Service
	validator *validator.Validate
}

func NewHandler(service *Service, v *validator.Validate) *Handler {
	return &Handler{service: service, validator: v}
}

func githubPrincipal(r *http.Request) error {
	_, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		return httpx.Unauthorized("unauthorized", "authorization required")
	}
	return nil
}

func (h *Handler) CreateConnection(w http.ResponseWriter, r *http.Request) {
	if err := githubPrincipal(r); err != nil {
		httpx.WriteError(w, err)
		return
	}
	principal, _ := platformauth.PrincipalFromContext(r.Context())
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

func (h *Handler) TestConnection(w http.ResponseWriter, r *http.Request) {
	if err := githubPrincipal(r); err != nil {
		httpx.WriteError(w, err)
		return
	}
	var req TestConnectionRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_request", err.Error()))
		return
	}
	item, err := h.service.TestConnection(r.Context(), req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) ListConnections(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListConnections(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}
func (h *Handler) GetConnection(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "connectionId"), "invalid_connection_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	item, err := h.service.GetConnection(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}
func (h *Handler) UpdateConnection(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "connectionId"), "invalid_connection_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	var req UpdateConnectionRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	item, err := h.service.UpdateConnection(r.Context(), id, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}
func (h *Handler) DeleteConnection(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "connectionId"), "invalid_connection_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.service.DeleteConnection(r.Context(), id); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteNoContent(w)
}
func (h *Handler) ImportUsers(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "connectionId"), "invalid_connection_id")
	if err != nil {
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
func (h *Handler) ImportRepos(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "connectionId"), "invalid_connection_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	item, err := h.service.ImportRepos(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}
func (h *Handler) ImportLanguages(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "connectionId"), "invalid_connection_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	item, err := h.service.ImportLanguages(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}
func (h *Handler) StartSync(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "connectionId"), "invalid_connection_id")
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
	item, err := h.service.StartSync(r.Context(), id, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"jobId": item.ID, "status": item.Status})
}
func (h *Handler) GetSyncJob(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "jobId"), "invalid_job_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	item, err := h.service.GetSyncJob(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}
func (h *Handler) ListMappings(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "connectionId"), "invalid_connection_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	items, err := h.service.ListMappings(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}
func (h *Handler) CreateMapping(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "connectionId"), "invalid_connection_id")
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
	if err := h.service.CreateMapping(r.Context(), id, req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteNoContent(w)
}
func (h *Handler) DeleteMapping(w http.ResponseWriter, r *http.Request) {
	connectionID, err := parseUUID(chi.URLParam(r, "connectionId"), "invalid_connection_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	mappingID, err := parseUUID(chi.URLParam(r, "mappingId"), "invalid_mapping_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.service.DeleteMapping(r.Context(), connectionID, mappingID); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteNoContent(w)
}
func (h *Handler) AutoMatchMappings(w http.ResponseWriter, r *http.Request) {
	connectionID, err := parseUUID(chi.URLParam(r, "connectionId"), "invalid_connection_id")
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
	item, err := h.service.AutoMatch(r.Context(), connectionID, req.Strategy)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}
func (h *Handler) ListGithubUsers(w http.ResponseWriter, r *http.Request) {
	var connectionID *uuid.UUID
	if raw := strings.TrimSpace(r.URL.Query().Get("connectionId")); raw != "" {
		id, err := parseUUID(raw, "invalid_connection_id")
		if err != nil {
			httpx.WriteError(w, err)
			return
		}
		connectionID = &id
	}
	limit, offset := parseLimitOffset(r)
	items, err := h.service.ListGitHubUsers(r.Context(), connectionID, limit, offset)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}
func (h *Handler) ListRepositories(w http.ResponseWriter, r *http.Request) {
	filters, err := parseRepositoryFilters(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	items, err := h.service.ListRepositories(r.Context(), filters)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}
func (h *Handler) GetRepository(w http.ResponseWriter, r *http.Request) {
	repoID, err := parseUUID(chi.URLParam(r, "repoId"), "invalid_repo_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	item, err := h.service.GetRepositoryDetail(r.Context(), repoID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}
func (h *Handler) GetRepositoryLanguages(w http.ResponseWriter, r *http.Request) {
	repoID, err := parseUUID(chi.URLParam(r, "repoId"), "invalid_repo_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	items, err := h.service.GetRepositoryLanguages(r.Context(), repoID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}
func (h *Handler) GetRepositoryContributors(w http.ResponseWriter, r *http.Request) {
	repoID, err := parseUUID(chi.URLParam(r, "repoId"), "invalid_repo_id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	items, err := h.service.GetRepositoryContributors(r.Context(), repoID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}
func (h *Handler) GetEmployeeProfile(w http.ResponseWriter, r *http.Request) {
	connectionID, employeeID, err := parseConnectionEmployee(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	item, err := h.service.GetEmployeeProfile(r.Context(), connectionID, employeeID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}
func (h *Handler) GetEmployeeLanguages(w http.ResponseWriter, r *http.Request) {
	connectionID, employeeID, err := parseConnectionEmployee(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	item, err := h.service.GetEmployeeLanguages(r.Context(), connectionID, employeeID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}
func (h *Handler) GetEmployeeStats(w http.ResponseWriter, r *http.Request) {
	connectionID, employeeID, err := parseConnectionEmployee(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	from, to := parsePeriod(r)
	item, err := h.service.GetEmployeeStats(r.Context(), connectionID, employeeID, from, to)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}
func (h *Handler) GetEmployeeActivity(w http.ResponseWriter, r *http.Request) {
	connectionID, employeeID, err := parseConnectionEmployee(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	from, to := parsePeriod(r)
	item, err := h.service.GetEmployeeActivity(r.Context(), connectionID, employeeID, from, to)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}
func (h *Handler) GetTeamAnalytics(w http.ResponseWriter, r *http.Request) {
	connectionID, err := parseConnectionIDQuery(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	from, to := parsePeriod(r)
	item, err := h.service.TeamAnalytics(r.Context(), connectionID, from, to)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}
func (h *Handler) GetLanguageAnalytics(w http.ResponseWriter, r *http.Request) {
	connectionID, err := parseConnectionIDQuery(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	item, err := h.service.LanguageAnalytics(r.Context(), connectionID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}
func (h *Handler) GetTopLanguages(w http.ResponseWriter, r *http.Request) {
	connectionID, err := parseConnectionIDQuery(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	item, err := h.service.TopLanguages(r.Context(), connectionID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}
func (h *Handler) GetRepositoryHealth(w http.ResponseWriter, r *http.Request) {
	connectionID, err := parseConnectionIDQuery(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	items, err := h.service.RepositoryHealth(r.Context(), connectionID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}
func (h *Handler) GetRepositoryOwnership(w http.ResponseWriter, r *http.Request) {
	connectionID, err := parseConnectionIDQuery(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	from, _ := parsePeriod(r)
	items, err := h.service.RepositoryOwnership(r.Context(), connectionID, from)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func parseUUID(raw, code string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return uuid.Nil, httpx.BadRequest(code, "invalid uuid")
	}
	return id, nil
}

func parseConnectionIDQuery(r *http.Request) (uuid.UUID, error) {
	return parseUUID(r.URL.Query().Get("connectionId"), "invalid_connection_id")
}

func parseConnectionEmployee(r *http.Request) (uuid.UUID, uuid.UUID, error) {
	connectionID, err := parseConnectionIDQuery(r)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	employeeID, err := parseUUID(chi.URLParam(r, "employeeUserId"), "invalid_employee_user_id")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return connectionID, employeeID, nil
}

func parseLimitOffset(r *http.Request) (int, int) {
	limit := 50
	offset := 0
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 && v <= 500 {
			limit = v
		}
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("offset")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v >= 0 {
			offset = v
		}
	}
	return limit, offset
}

func parseRepositoryFilters(r *http.Request) (RepositoryFilters, error) {
	limit, offset := parseLimitOffset(r)
	filters := RepositoryFilters{Limit: limit, Offset: offset, Owner: r.URL.Query().Get("owner"), Language: r.URL.Query().Get("language"), Visibility: r.URL.Query().Get("visibility")}
	if raw := strings.TrimSpace(r.URL.Query().Get("connectionId")); raw != "" {
		id, err := parseUUID(raw, "invalid_connection_id")
		if err != nil {
			return filters, err
		}
		filters.ConnectionID = &id
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("employeeUserId")); raw != "" {
		id, err := parseUUID(raw, "invalid_employee_user_id")
		if err != nil {
			return filters, err
		}
		filters.EmployeeUserID = &id
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("archived")); raw != "" {
		val := raw == "true" || raw == "1"
		filters.Archived = &val
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("fork")); raw != "" {
		val := raw == "true" || raw == "1"
		filters.Fork = &val
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("activeSince")); raw != "" {
		if t, err := time.Parse("2006-01-02", raw); err == nil {
			filters.ActiveSince = &t
		}
	}
	return filters, nil
}

func parsePeriod(r *http.Request) (time.Time, time.Time) {
	to := time.Now().UTC()
	from := to.AddDate(0, 0, -30)
	if raw := strings.TrimSpace(r.URL.Query().Get("from")); raw != "" {
		if t, err := time.Parse("2006-01-02", raw); err == nil {
			from = t
		}
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("to")); raw != "" {
		if t, err := time.Parse("2006-01-02", raw); err == nil {
			to = t
		}
	}
	return from, to
}

func shouldMatch(strategy, email, localPart, domain string, gh GitHubUser) bool {
	switch strategy {
	case "email":
		return gh.Email != nil && strings.EqualFold(strings.TrimSpace(email), strings.TrimSpace(*gh.Email))
	case "login":
		return strings.EqualFold(localPart, gh.Login)
	case "domain":
		if gh.Email == nil {
			return false
		}
		ghLocal, ghDomain := splitEmail(*gh.Email)
		return strings.EqualFold(domain, ghDomain) && strings.EqualFold(localPart, ghLocal)
	default:
		return false
	}
}

func splitEmail(email string) (string, string) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(email)), "@")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func defaultGitHubBaseURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "https://api.github.com"
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "https://api.github.com"
	}
	return strings.TrimRight(raw, "/")
}

func rawJSON(v []byte) string {
	if len(bytes.TrimSpace(v)) == 0 {
		return "{}"
	}
	return string(v)
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func stringPtr(v string) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return &v
}

func lastN(v string, n int) string {
	v = strings.TrimSpace(v)
	if len(v) <= n {
		return v
	}
	return v[len(v)-n:]
}

func stringFromAny(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	default:
		if t == nil {
			return ""
		}
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

func optionalString(v any) *string { return stringPtr(stringFromAny(v)) }

func boolFromAny(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return strings.EqualFold(strings.TrimSpace(t), "true") || strings.TrimSpace(t) == "1"
	case float64:
		return t != 0
	default:
		return false
	}
}

func intFromAny(v any) int {
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case string:
		x, _ := strconv.Atoi(strings.TrimSpace(t))
		return x
	default:
		return 0
	}
}

func intPtrFromAny(v any) *int {
	x := intFromAny(v)
	if x == 0 && stringFromAny(v) == "" {
		return nil
	}
	return &x
}

func int64FromAny(v any) int64 {
	switch t := v.(type) {
	case int64:
		return t
	case int:
		return int64(t)
	case float64:
		return int64(t)
	case json.Number:
		x, _ := t.Int64()
		return x
	case string:
		x, _ := strconv.ParseInt(strings.TrimSpace(t), 10, 64)
		return x
	default:
		return 0
	}
}

func timePtrFromAny(v any) *time.Time {
	s := stringFromAny(v)
	if s == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05Z07:00"} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}

func percentOf(part, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) * 100 / float64(total)
}

func formatDecimal(v float64) string {
	return strconv.FormatFloat(math.Round(v*100)/100, 'f', -1, 64)
}

func intOrZero(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func engineeringScore(stats EmployeeStats) float64 {
	freshness, _ := strconv.ParseFloat(stats.AvgRepoFreshnessDays, 64)
	freshnessScore := math.Max(0, 100-math.Min(100, freshness*4))
	repoActivityScore := math.Min(100, float64(stats.ActiveRepositoriesCount)*18)
	ownershipScore := math.Min(100, float64(stats.RepositoriesCount)*8)
	deliveryScore := math.Min(100, float64(stats.Commits)*2+float64(stats.MergedPRs)*8+float64(stats.ReviewedPRs)*4)
	return 0.35*freshnessScore + 0.30*repoActivityScore + 0.15*ownershipScore + 0.20*deliveryScore
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
