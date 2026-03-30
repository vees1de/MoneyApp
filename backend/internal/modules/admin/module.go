package admin

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"moneyapp/backend/internal/modules/identity"
	"moneyapp/backend/internal/modules/org"
	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/clock"
	"moneyapp/backend/internal/platform/db"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CreateUserRequest struct {
	Email         string     `json:"email" validate:"required,email"`
	Status        string     `json:"status" validate:"omitempty,oneof=active invited blocked deleted"`
	FirstName     string     `json:"first_name" validate:"required"`
	LastName      string     `json:"last_name" validate:"required"`
	MiddleName    *string    `json:"middle_name,omitempty"`
	PositionTitle *string    `json:"position_title,omitempty"`
	DepartmentID  *uuid.UUID `json:"department_id,omitempty"`
	Timezone      *string    `json:"timezone,omitempty"`
	RoleCodes     []string   `json:"role_codes,omitempty"`
	Password      *string    `json:"password,omitempty"`
}

type UpdateUserRequest struct {
	Email         *string    `json:"email,omitempty"`
	Status        *string    `json:"status,omitempty"`
	FirstName     *string    `json:"first_name,omitempty"`
	LastName      *string    `json:"last_name,omitempty"`
	MiddleName    *string    `json:"middle_name,omitempty"`
	PositionTitle *string    `json:"position_title,omitempty"`
	DepartmentID  *uuid.UUID `json:"department_id,omitempty"`
	Timezone      *string    `json:"timezone,omitempty"`
	OutlookEmail  *string    `json:"outlook_email,omitempty"`
}

type AssignRoleRequest struct {
	RoleCode string `json:"role_code,omitempty"`
}

type Service struct {
	db         *sql.DB
	repo       *identity.Repository
	orgService *org.Service
	clock      clock.Clock
}

func NewService(database *sql.DB, repo *identity.Repository, orgService *org.Service, appClock clock.Clock) *Service {
	return &Service{
		db:         database,
		repo:       repo,
		orgService: orgService,
		clock:      appClock,
	}
}

func (s *Service) ListUsers(ctx context.Context) ([]identity.UserView, error) {
	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]identity.UserView, 0, len(users))
	for _, user := range users {
		principal, err := s.repo.BuildPrincipal(ctx, user.ID, uuid.Nil, s.orgService)
		if err != nil {
			return nil, err
		}
		view := identity.UserView{
			User:        user,
			Roles:       principal.RoleCodes,
			Permissions: principal.PermissionCodes,
		}
		if profile, err := s.orgService.GetProfileByUserID(ctx, user.ID); err == nil {
			view.EmployeeProfile = &profile
		}
		items = append(items, view)
	}

	return items, nil
}

func (s *Service) GetUser(ctx context.Context, userID uuid.UUID) (identity.UserView, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return identity.UserView{}, httpx.NotFound("user_not_found", "user not found")
		}
		return identity.UserView{}, err
	}

	principal, err := s.repo.BuildPrincipal(ctx, userID, uuid.Nil, s.orgService)
	if err != nil {
		return identity.UserView{}, err
	}

	view := identity.UserView{
		User:        user,
		Roles:       principal.RoleCodes,
		Permissions: principal.PermissionCodes,
	}
	if profile, err := s.orgService.GetProfileByUserID(ctx, userID); err == nil {
		view.EmployeeProfile = &profile
	}

	return view, nil
}

func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) (identity.UserView, error) {
	var response identity.UserView
	err := db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if _, _, err := s.repo.GetUserByEmail(ctx, req.Email, tx); err == nil {
			return httpx.Conflict("email_taken", "user with this email already exists")
		} else if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		status := strings.TrimSpace(req.Status)
		if status == "" {
			if req.Password == nil || strings.TrimSpace(*req.Password) == "" {
				status = "invited"
			} else {
				status = "active"
			}
		}

		now := s.clock.Now()
		user := identity.User{
			ID:              uuid.New(),
			Email:           strings.TrimSpace(strings.ToLower(req.Email)),
			Status:          status,
			IsEmailVerified: false,
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		var passwordHash *string
		if req.Password != nil && strings.TrimSpace(*req.Password) != "" {
			hash, err := platformauth.HashPassword(*req.Password)
			if err != nil {
				return err
			}
			passwordHash = &hash
		}

		if err := s.repo.CreateUser(ctx, user, passwordHash, tx); err != nil {
			return err
		}

		if _, err := s.orgService.CreateDefaultProfile(ctx, org.CreateProfileInput{
			UserID:        user.ID,
			FirstName:     req.FirstName,
			LastName:      req.LastName,
			MiddleName:    req.MiddleName,
			PositionTitle: req.PositionTitle,
			DepartmentID:  req.DepartmentID,
			Timezone:      req.Timezone,
		}, tx); err != nil {
			return err
		}

		roleCodes := req.RoleCodes
		if len(roleCodes) == 0 {
			roleCodes = []string{"employee"}
		}
		for _, roleCode := range roleCodes {
			if err := s.repo.AssignRoleByCode(ctx, user.ID, roleCode, nil, nil, tx); err != nil {
				return err
			}
		}

		principal, err := s.repo.BuildPrincipal(ctx, user.ID, uuid.Nil, s.orgService, tx)
		if err != nil {
			return err
		}
		response = identity.UserView{
			User:        user,
			Roles:       principal.RoleCodes,
			Permissions: principal.PermissionCodes,
		}
		if profile, err := s.orgService.GetProfileByUserID(ctx, user.ID, tx); err == nil {
			response.EmployeeProfile = &profile
		}

		return nil
	})

	return response, err
}

func (s *Service) UpdateUser(ctx context.Context, userID uuid.UUID, req UpdateUserRequest) (identity.UserView, error) {
	err := db.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		user, err := s.repo.GetUserByID(ctx, userID, tx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return httpx.NotFound("user_not_found", "user not found")
			}
			return err
		}

		if req.Email != nil {
			user.Email = strings.TrimSpace(strings.ToLower(*req.Email))
		}
		if req.Status != nil {
			user.Status = *req.Status
		}
		user.UpdatedAt = s.clock.Now()

		if err := s.repo.UpdateUser(ctx, user, nil, tx); err != nil {
			return err
		}

		if _, err := s.orgService.UpdateProfile(ctx, userID, org.UpdateProfileInput{
			FirstName:     req.FirstName,
			LastName:      req.LastName,
			MiddleName:    req.MiddleName,
			PositionTitle: req.PositionTitle,
			DepartmentID:  req.DepartmentID,
			Timezone:      req.Timezone,
			OutlookEmail:  req.OutlookEmail,
		}, tx); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		return nil
	})
	if err != nil {
		return identity.UserView{}, err
	}

	return s.GetUser(ctx, userID)
}

func (s *Service) AssignRole(ctx context.Context, userID uuid.UUID, roleCode string) error {
	return s.repo.AssignRoleByCode(ctx, userID, roleCode, nil, nil)
}

func (s *Service) RemoveRole(ctx context.Context, userID, roleID uuid.UUID) error {
	return s.repo.RemoveRoleByID(ctx, userID, roleID)
}

func (s *Service) ListRoles(ctx context.Context) ([]identity.Role, error) {
	return s.repo.ListRoles(ctx)
}

func (s *Service) ListPermissions(ctx context.Context) ([]identity.Permission, error) {
	return s.repo.ListPermissions(ctx)
}

type Handler struct {
	service   *Service
	validator *validator.Validate
}

func NewHandler(service *Service, validator *validator.Validate) *Handler {
	return &Handler{service: service, validator: validator}
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListUsers(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}

	user, err := h.service.CreateUser(r.Context(), req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, user)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_user_id", "invalid user id"))
		return
	}

	user, err := h.service.GetUser(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_user_id", "invalid user id"))
		return
	}

	var req UpdateUserRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	user, err := h.service.UpdateUser(r.Context(), userID, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) AssignRole(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_user_id", "invalid user id"))
		return
	}

	var req AssignRoleRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if strings.TrimSpace(req.RoleCode) == "" {
		httpx.WriteError(w, httpx.BadRequest("validation_error", "role_code is required"))
		return
	}

	if err := h.service.AssignRole(r.Context(), userID, req.RoleCode); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) RemoveRole(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_user_id", "invalid user id"))
		return
	}
	roleID, err := uuid.Parse(chi.URLParam(r, "roleId"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_role_id", "invalid role id"))
		return
	}

	if err := h.service.RemoveRole(r.Context(), userID, roleID); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListRoles(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListPermissions(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}
