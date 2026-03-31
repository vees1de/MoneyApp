package identity

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"moneyapp/backend/internal/modules/org"
	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/clock"
	"moneyapp/backend/internal/platform/db"
	"moneyapp/backend/internal/platform/events"
	"moneyapp/backend/internal/platform/httpx"
	"moneyapp/backend/internal/platform/outbox"
	"moneyapp/backend/internal/platform/worker"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID  `json:"id"`
	Email           string     `json:"email"`
	Status          string     `json:"status"`
	IsEmailVerified bool       `json:"is_email_verified"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

type UserView struct {
	User
	Roles           []string             `json:"roles"`
	Permissions     []string             `json:"permissions"`
	EmployeeProfile *org.EmployeeProfile `json:"employee_profile,omitempty"`
}

type Role struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	IsSystem    bool      `json:"is_system"`
}

type Permission struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Module      string    `json:"module"`
	Action      string    `json:"action"`
	Description *string   `json:"description,omitempty"`
}

type Session struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	RefreshTokenHash string
	UserAgent        *string
	IP               *string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	CreatedAt        time.Time
}

type PasswordReset struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type AuthResponse struct {
	User   UserView       `json:"user"`
	Tokens Tokens         `json:"tokens"`
	Meta   map[string]any `json:"meta,omitempty"`
}

type MeResponse struct {
	User UserView `json:"user"`
}

type RegisterRequest struct {
	Email         string  `json:"email" validate:"required,email"`
	Password      string  `json:"password" validate:"required,min=8"`
	FirstName     string  `json:"first_name" validate:"required"`
	LastName      string  `json:"last_name" validate:"required"`
	MiddleName    *string `json:"middle_name,omitempty"`
	PositionTitle *string `json:"position_title,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
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

func (r *Repository) CreateUser(ctx context.Context, user User, passwordHash *string, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into users (
			id, email, password_hash, status, is_email_verified, last_login_at, created_at, updated_at, deleted_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, user.ID, user.Email, passwordHash, user.Status, user.IsEmailVerified, user.LastLoginAt, user.CreatedAt, user.UpdatedAt, user.DeletedAt)
	return err
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string, exec ...db.DBTX) (User, *string, error) {
	normalized := strings.TrimSpace(strings.ToLower(email))
	var user User
	var passwordHash *string
	err := r.base(exec...).QueryRowContext(ctx, `
		select id, email, password_hash, status, is_email_verified, last_login_at, created_at, updated_at, deleted_at
		from users
		where lower(email::text) = $1
	`, normalized).Scan(
		&user.ID, &user.Email, &passwordHash, &user.Status, &user.IsEmailVerified, &user.LastLoginAt,
		&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)
	return user, passwordHash, err
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID, exec ...db.DBTX) (User, error) {
	var user User
	err := r.base(exec...).QueryRowContext(ctx, `
		select id, email, status, is_email_verified, last_login_at, created_at, updated_at, deleted_at
		from users
		where id = $1
	`, id).Scan(
		&user.ID, &user.Email, &user.Status, &user.IsEmailVerified, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)
	return user, err
}

func (r *Repository) UpdateUser(ctx context.Context, user User, passwordHash *string, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update users
		set email = $2,
		    password_hash = coalesce($3, password_hash),
		    status = $4,
		    is_email_verified = $5,
		    last_login_at = $6,
		    updated_at = $7,
		    deleted_at = $8
		where id = $1
	`, user.ID, user.Email, passwordHash, user.Status, user.IsEmailVerified, user.LastLoginAt, user.UpdatedAt, user.DeletedAt)
	return err
}

func (r *Repository) TouchLastLogin(ctx context.Context, userID uuid.UUID, when time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update users
		set last_login_at = $2, updated_at = $2
		where id = $1
	`, userID, when)
	return err
}

func (r *Repository) ListUsers(ctx context.Context, exec ...db.DBTX) ([]User, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select id, email, status, is_email_verified, last_login_at, created_at, updated_at, deleted_at
		from users
		order by created_at desc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Email, &user.Status, &user.IsEmailVerified, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt); err != nil {
			return nil, err
		}
		items = append(items, user)
	}
	return items, rows.Err()
}

func (r *Repository) ListRoles(ctx context.Context, exec ...db.DBTX) ([]Role, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select id, code, name, description, is_system
		from roles
		order by code asc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Code, &role.Name, &role.Description, &role.IsSystem); err != nil {
			return nil, err
		}
		items = append(items, role)
	}
	return items, rows.Err()
}

func (r *Repository) ListPermissions(ctx context.Context, exec ...db.DBTX) ([]Permission, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select id, code, module, action, description
		from permissions
		order by code asc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Permission
	for rows.Next() {
		var permission Permission
		if err := rows.Scan(&permission.ID, &permission.Code, &permission.Module, &permission.Action, &permission.Description); err != nil {
			return nil, err
		}
		items = append(items, permission)
	}
	return items, rows.Err()
}

func (r *Repository) FindUserIDByRoleCode(ctx context.Context, roleCode string, exec ...db.DBTX) (*uuid.UUID, error) {
	var userID uuid.UUID
	err := r.base(exec...).QueryRowContext(ctx, `
		select ur.user_id
		from user_roles ur
		join roles r on r.id = ur.role_id
		where r.code = $1
		order by ur.created_at asc
		limit 1
	`, roleCode).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &userID, nil
}

func (r *Repository) AssignRoleByCode(ctx context.Context, userID uuid.UUID, roleCode string, scopeType *string, scopeID *uuid.UUID, exec ...db.DBTX) error {
	scopeValue := "global"
	if scopeType != nil && *scopeType != "" {
		scopeValue = *scopeType
	}

	_, err := r.base(exec...).ExecContext(ctx, `
		insert into user_roles (user_id, role_id, scope_type, scope_id, created_at)
		select $1, id, $3, $4, $5
		from roles
		where code = $2
		on conflict do nothing
	`, userID, roleCode, scopeValue, scopeID, time.Now().UTC())
	return err
}

func (r *Repository) RemoveRoleByID(ctx context.Context, userID, roleID uuid.UUID, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		delete from user_roles
		where user_id = $1 and role_id = $2
	`, userID, roleID)
	return err
}

func (r *Repository) CreateSession(ctx context.Context, session Session, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into sessions (id, user_id, refresh_token_hash, user_agent, ip, expires_at, revoked_at, created_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
	`, session.ID, session.UserID, session.RefreshTokenHash, session.UserAgent, session.IP, session.ExpiresAt, session.RevokedAt, session.CreatedAt)
	return err
}

func (r *Repository) GetSessionByRefreshHash(ctx context.Context, hash string, exec ...db.DBTX) (Session, error) {
	var item Session
	err := r.base(exec...).QueryRowContext(ctx, `
		select id, user_id, refresh_token_hash, user_agent, host(ip), expires_at, revoked_at, created_at
		from sessions
		where refresh_token_hash = $1
	`, hash).Scan(&item.ID, &item.UserID, &item.RefreshTokenHash, &item.UserAgent, &item.IP, &item.ExpiresAt, &item.RevokedAt, &item.CreatedAt)
	return item, err
}

func (r *Repository) RotateSession(ctx context.Context, sessionID uuid.UUID, hash string, userAgent, ip *string, expiresAt time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update sessions
		set refresh_token_hash = $2,
		    user_agent = $3,
		    ip = $4,
		    expires_at = $5,
		    revoked_at = null
		where id = $1
	`, sessionID, hash, userAgent, ip, expiresAt)
	return err
}

func (r *Repository) RevokeSessionByHash(ctx context.Context, hash string, when time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update sessions
		set revoked_at = $2
		where refresh_token_hash = $1 and revoked_at is null
	`, hash, when)
	return err
}

func (r *Repository) CreatePasswordReset(ctx context.Context, item PasswordReset, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		insert into password_resets (id, user_id, token_hash, expires_at, used_at, created_at)
		values ($1, $2, $3, $4, $5, $6)
	`, item.ID, item.UserID, item.TokenHash, item.ExpiresAt, item.UsedAt, item.CreatedAt)
	return err
}

func (r *Repository) GetPasswordResetByHash(ctx context.Context, hash string, exec ...db.DBTX) (PasswordReset, error) {
	var item PasswordReset
	err := r.base(exec...).QueryRowContext(ctx, `
		select id, user_id, token_hash, expires_at, used_at, created_at
		from password_resets
		where token_hash = $1
	`, hash).Scan(&item.ID, &item.UserID, &item.TokenHash, &item.ExpiresAt, &item.UsedAt, &item.CreatedAt)
	return item, err
}

func (r *Repository) MarkPasswordResetUsed(ctx context.Context, id uuid.UUID, when time.Time, exec ...db.DBTX) error {
	_, err := r.base(exec...).ExecContext(ctx, `
		update password_resets
		set used_at = $2
		where id = $1
	`, id, when)
	return err
}

func (r *Repository) BuildPrincipal(ctx context.Context, userID, sessionID uuid.UUID, orgService *org.Service, exec ...db.DBTX) (platformauth.Principal, error) {
	rows, err := r.base(exec...).QueryContext(ctx, `
		select distinct r.code
		from user_roles ur
		join roles r on r.id = ur.role_id
		where ur.user_id = $1
		order by r.code asc
	`, userID)
	if err != nil {
		return platformauth.Principal{}, err
	}
	defer rows.Close()

	roleCodes := make([]string, 0, 4)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return platformauth.Principal{}, err
		}
		roleCodes = append(roleCodes, code)
	}
	if err := rows.Err(); err != nil {
		return platformauth.Principal{}, err
	}

	permissionRows, err := r.base(exec...).QueryContext(ctx, `
		select distinct p.code
		from user_roles ur
		join role_permissions rp on rp.role_id = ur.role_id
		join permissions p on p.id = rp.permission_id
		where ur.user_id = $1
		order by p.code asc
	`, userID)
	if err != nil {
		return platformauth.Principal{}, err
	}
	defer permissionRows.Close()

	permissionCodes := make([]string, 0, 8)
	for permissionRows.Next() {
		var code string
		if err := permissionRows.Scan(&code); err != nil {
			return platformauth.Principal{}, err
		}
		permissionCodes = append(permissionCodes, code)
	}
	if err := permissionRows.Err(); err != nil {
		return platformauth.Principal{}, err
	}

	var employeeProfileID *uuid.UUID
	var departmentID *uuid.UUID
	if profile, err := orgService.GetProfileByUserID(ctx, userID, exec...); err == nil {
		employeeProfileID = &profile.ID
		departmentID = profile.DepartmentID
	} else if !errors.Is(err, sql.ErrNoRows) {
		return platformauth.Principal{}, err
	}

	return platformauth.Principal{
		UserID:            userID,
		SessionID:         sessionID,
		RoleCodes:         roleCodes,
		PermissionCodes:   platformauth.WithImplicitPermissions(permissionCodes),
		EmployeeProfileID: employeeProfileID,
		DepartmentID:      departmentID,
	}, nil
}

type Service struct {
	db         *sql.DB
	repo       *Repository
	orgService *org.Service
	outbox     *outbox.Service
	queue      *worker.Queue
	jwt        *platformauth.JWTManager
	clock      clock.Clock
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewService(database *sql.DB, repo *Repository, orgService *org.Service, outboxService *outbox.Service, queue *worker.Queue, jwt *platformauth.JWTManager, appClock clock.Clock, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		db:         database,
		repo:       repo,
		orgService: orgService,
		outbox:     outboxService,
		queue:      queue,
		jwt:        jwt,
		clock:      appClock,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest, meta SessionMeta) (AuthResponse, error) {
	var response AuthResponse

	err := db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if _, _, err := s.repo.GetUserByEmail(ctx, req.Email, tx); err == nil {
			return httpx.Conflict("email_taken", "user with this email already exists")
		} else if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		now := s.clock.Now()
		passwordHash, err := platformauth.HashPassword(req.Password)
		if err != nil {
			return err
		}

		user := User{
			ID:              uuid.New(),
			Email:           strings.TrimSpace(strings.ToLower(req.Email)),
			Status:          "active",
			IsEmailVerified: false,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := s.repo.CreateUser(ctx, user, &passwordHash, tx); err != nil {
			return err
		}

		if _, err := s.orgService.CreateDefaultProfile(ctx, org.CreateProfileInput{
			UserID:        user.ID,
			FirstName:     req.FirstName,
			LastName:      req.LastName,
			MiddleName:    req.MiddleName,
			PositionTitle: req.PositionTitle,
		}, tx); err != nil {
			return err
		}

		if err := s.repo.AssignRoleByCode(ctx, user.ID, "employee", nil, nil, tx); err != nil {
			return err
		}

		principal, tokens, session, err := s.issueTokens(ctx, user.ID, meta, tx)
		if err != nil {
			return err
		}

		if err := s.outbox.Publish(ctx, tx, events.Message{
			Topic:      "identity",
			EventType:  "identity.user.registered",
			EntityType: "user",
			EntityID:   user.ID,
			Payload: map[string]any{
				"email": user.Email,
			},
			OccurredAt: now,
		}); err != nil {
			return err
		}

		view, err := s.userView(ctx, principal, tx)
		if err != nil {
			return err
		}

		response = AuthResponse{
			User:   view,
			Tokens: tokens,
			Meta: map[string]any{
				"created":    true,
				"session_id": session.ID.String(),
			},
		}
		return nil
	})

	return response, err
}

func (s *Service) Login(ctx context.Context, req LoginRequest, meta SessionMeta) (AuthResponse, error) {
	user, passwordHash, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AuthResponse{}, httpx.Unauthorized("invalid_credentials", "invalid email or password")
		}
		return AuthResponse{}, err
	}

	if passwordHash == nil || platformauth.ComparePassword(*passwordHash, req.Password) != nil {
		return AuthResponse{}, httpx.Unauthorized("invalid_credentials", "invalid email or password")
	}
	if user.Status != "active" && user.Status != "invited" {
		return AuthResponse{}, httpx.Forbidden("user_blocked", "user is not active")
	}

	now := s.clock.Now()
	if err := s.repo.TouchLastLogin(ctx, user.ID, now); err != nil {
		return AuthResponse{}, err
	}

	principal, tokens, _, err := s.issueTokens(ctx, user.ID, meta)
	if err != nil {
		return AuthResponse{}, err
	}
	view, err := s.userView(ctx, principal)
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{
		User:   view,
		Tokens: tokens,
	}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string, meta SessionMeta) (AuthResponse, error) {
	hash := platformauth.HashToken(refreshToken)
	session, err := s.repo.GetSessionByRefreshHash(ctx, hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AuthResponse{}, httpx.Unauthorized("invalid_session", "refresh token is invalid")
		}
		return AuthResponse{}, err
	}

	now := s.clock.Now()
	if session.RevokedAt != nil || session.ExpiresAt.Before(now) {
		return AuthResponse{}, httpx.Unauthorized("expired_session", "refresh session expired")
	}

	nextRefresh, err := platformauth.NewOpaqueToken()
	if err != nil {
		return AuthResponse{}, err
	}
	session.RefreshTokenHash = platformauth.HashToken(nextRefresh)
	session.UserAgent = meta.UserAgent
	session.IP = meta.IP
	session.ExpiresAt = now.Add(s.refreshTTL)

	if err := s.repo.RotateSession(ctx, session.ID, session.RefreshTokenHash, session.UserAgent, session.IP, session.ExpiresAt); err != nil {
		return AuthResponse{}, err
	}

	principal, err := s.repo.BuildPrincipal(ctx, session.UserID, session.ID, s.orgService)
	if err != nil {
		return AuthResponse{}, err
	}
	accessToken, err := s.jwt.SignPrincipalToken(principal, now)
	if err != nil {
		return AuthResponse{}, err
	}

	view, err := s.userView(ctx, principal)
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{
		User: view,
		Tokens: Tokens{
			AccessToken:  accessToken,
			RefreshToken: nextRefresh,
			ExpiresIn:    int64(s.accessTTL.Seconds()),
		},
	}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	return s.repo.RevokeSessionByHash(ctx, platformauth.HashToken(refreshToken), s.clock.Now())
}

func (s *Service) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) error {
	user, _, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	rawToken, err := platformauth.NewOpaqueToken()
	if err != nil {
		return err
	}

	now := s.clock.Now()
	item := PasswordReset{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: platformauth.HashToken(rawToken),
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
	}
	if err := s.repo.CreatePasswordReset(ctx, item); err != nil {
		return err
	}

	key := "password-reset:" + user.ID.String()
	return s.queue.Enqueue(ctx, s.db, "notifications", "send_password_reset", map[string]any{
		"user_id": user.ID,
		"email":   user.Email,
		"token":   rawToken,
	}, &key, now)
}

func (s *Service) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	item, err := s.repo.GetPasswordResetByHash(ctx, platformauth.HashToken(req.Token))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return httpx.BadRequest("invalid_reset_token", "password reset token is invalid")
		}
		return err
	}

	now := s.clock.Now()
	if item.UsedAt != nil || item.ExpiresAt.Before(now) {
		return httpx.BadRequest("expired_reset_token", "password reset token expired")
	}

	passwordHash, err := platformauth.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	return db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		user, err := s.repo.GetUserByID(ctx, item.UserID, tx)
		if err != nil {
			return err
		}
		user.Status = "active"
		user.IsEmailVerified = true
		user.UpdatedAt = now

		if err := s.repo.UpdateUser(ctx, user, &passwordHash, tx); err != nil {
			return err
		}
		return s.repo.MarkPasswordResetUsed(ctx, item.ID, now, tx)
	})
}

func (s *Service) Me(ctx context.Context, principal platformauth.Principal) (UserView, error) {
	return s.userView(ctx, principal)
}

func (s *Service) issueTokens(ctx context.Context, userID uuid.UUID, meta SessionMeta, exec ...db.DBTX) (platformauth.Principal, Tokens, Session, error) {
	now := s.clock.Now()
	rawRefresh, err := platformauth.NewOpaqueToken()
	if err != nil {
		return platformauth.Principal{}, Tokens{}, Session{}, err
	}

	session := Session{
		ID:               uuid.New(),
		UserID:           userID,
		RefreshTokenHash: platformauth.HashToken(rawRefresh),
		UserAgent:        meta.UserAgent,
		IP:               meta.IP,
		ExpiresAt:        now.Add(s.refreshTTL),
		CreatedAt:        now,
	}
	if err := s.repo.CreateSession(ctx, session, exec...); err != nil {
		return platformauth.Principal{}, Tokens{}, Session{}, err
	}

	principal, err := s.repo.BuildPrincipal(ctx, userID, session.ID, s.orgService, exec...)
	if err != nil {
		return platformauth.Principal{}, Tokens{}, Session{}, err
	}

	accessToken, err := s.jwt.SignPrincipalToken(principal, now)
	if err != nil {
		return platformauth.Principal{}, Tokens{}, Session{}, err
	}

	return principal, Tokens{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, session, nil
}

func (s *Service) userView(ctx context.Context, principal platformauth.Principal, exec ...db.DBTX) (UserView, error) {
	user, err := s.repo.GetUserByID(ctx, principal.UserID, exec...)
	if err != nil {
		return UserView{}, err
	}

	var profile *org.EmployeeProfile
	if item, err := s.orgService.GetProfileByUserID(ctx, principal.UserID, exec...); err == nil {
		profile = &item
	} else if !errors.Is(err, sql.ErrNoRows) {
		return UserView{}, err
	}

	return UserView{
		User:            user,
		Roles:           principal.RoleCodes,
		Permissions:     principal.PermissionCodes,
		EmployeeProfile: profile,
	}, nil
}

type SessionMeta struct {
	UserAgent *string
	IP        *string
}

func SessionMetaFromRequest(r *http.Request) SessionMeta {
	meta := SessionMeta{
		UserAgent: nil,
		IP:        nil,
	}

	if ua := strings.TrimSpace(r.UserAgent()); ua != "" {
		meta.UserAgent = &ua
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		host = strings.TrimSpace(r.RemoteAddr)
	}
	if ip := net.ParseIP(host); ip != nil {
		value := ip.String()
		meta.IP = &value
	}

	return meta
}

type Handler struct {
	service   *Service
	validator *validator.Validate
}

func NewHandler(service *Service, validator *validator.Validate) *Handler {
	return &Handler{service: service, validator: validator}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}

	response, err := h.service.Register(r.Context(), req, SessionMetaFromRequest(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, response)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}

	response, err := h.service.Login(r.Context(), req, SessionMetaFromRequest(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}

	response, err := h.service.Refresh(r.Context(), req.RefreshToken, SessionMetaFromRequest(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}

	if err := h.service.Logout(r.Context(), req.RefreshToken); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}

	if err := h.service.ForgotPassword(r.Context(), req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}

	if err := h.service.ResetPassword(r.Context(), req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}

	user, err := h.service.Me(r.Context(), principal)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, MeResponse{User: user})
}
