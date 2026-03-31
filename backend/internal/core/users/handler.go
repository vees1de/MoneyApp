package users

import (
	"errors"
	"io"
	"net/http"
	"strings"

	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

const maxAvatarSizeBytes = 5 << 20

type Handler struct {
	service   *Service
	validator *validator.Validate
}

func NewHandler(service *Service, validator *validator.Validate) *Handler {
	return &Handler{
		service:   service,
		validator: validator,
	}
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}

	response, err := h.service.GetProfile(r.Context(), principal.UserID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}

	var request UpdateProfileRequest
	if err := httpx.DecodeJSON(r, &request); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(request); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}

	response, err := h.service.UpdateProfile(r.Context(), principal.UserID, request)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarSizeBytes+1024)
	if err := r.ParseMultipartForm(maxAvatarSizeBytes + 1024); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "too large") {
			httpx.WriteError(w, httpx.BadRequest("file_too_large", "avatar size must be up to 5MB"))
			return
		}
		httpx.WriteError(w, httpx.BadRequest("invalid_multipart", err.Error()))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			httpx.WriteError(w, httpx.BadRequest("file_required", "avatar file is required"))
			return
		}
		httpx.WriteError(w, httpx.BadRequest("invalid_multipart", err.Error()))
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if len(content) == 0 {
		httpx.WriteError(w, httpx.BadRequest("file_required", "avatar file is required"))
		return
	}
	if len(content) > maxAvatarSizeBytes {
		httpx.WriteError(w, httpx.BadRequest("file_too_large", "avatar size must be up to 5MB"))
		return
	}

	contentType := http.DetectContentType(content)
	response, err := h.service.UploadAvatar(r.Context(), principal.UserID, AvatarUpload{
		OriginalName: header.Filename,
		ContentType:  contentType,
		Content:      content,
		BaseURL:      requestBaseURL(r),
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) ListProfileRoles(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListProfileRoles(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, ProfileRolesResponse{Items: items})
}

func (h *Handler) ListDevelopmentTeams(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}

	items, err := h.service.ListDevelopmentTeams(r.Context(), principal.UserID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, DevelopmentTeamsResponse{Items: items})
}

func (h *Handler) ListAvailableDevelopmentTeams(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}

	items, err := h.service.ListAvailableDevelopmentTeams(r.Context(), principal.UserID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, DevelopmentTeamsResponse{Items: items})
}

func (h *Handler) CreateDevelopmentTeam(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}

	var request CreateDevelopmentTeamRequest
	if err := httpx.DecodeJSON(r, &request); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(request); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}

	response, err := h.service.CreateDevelopmentTeam(r.Context(), principal.UserID, request)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, response)
}

func (h *Handler) JoinDevelopmentTeam(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}

	teamID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_team_id", "invalid team id"))
		return
	}

	response, err := h.service.JoinDevelopmentTeam(r.Context(), principal.UserID, teamID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) LeaveCurrentDevelopmentTeam(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}

	response, err := h.service.LeaveCurrentDevelopmentTeam(r.Context(), principal.UserID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

func requestBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if value := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); value != "" {
		scheme = value
	}

	host := strings.TrimSpace(r.Host)
	if value := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); value != "" {
		host = value
	}

	return scheme + "://" + host
}
