package auth

import (
	"net/http"

	"moneyapp/backend/internal/core/sessions"
	"moneyapp/backend/internal/core/users"
	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/go-playground/validator/v10"
)

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

func (h *Handler) TelegramLogin(w http.ResponseWriter, r *http.Request) {
	var request TelegramLoginRequest
	if err := httpx.DecodeJSON(r, &request); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(request); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}

	response, err := h.service.LoginWithTelegram(r.Context(), request, sessionMetaFromRequest(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) YandexLogin(w http.ResponseWriter, r *http.Request) {
	var request YandexLoginRequest
	if err := httpx.DecodeJSON(r, &request); err != nil {
		httpx.WriteError(w, err)
		return
	}

	response, err := h.service.LoginWithYandex(r.Context(), request, sessionMetaFromRequest(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var request RefreshRequest
	if err := httpx.DecodeJSON(r, &request); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(request); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}

	response, err := h.service.Refresh(r.Context(), request.RefreshToken, sessionMetaFromRequest(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var request LogoutRequest
	if err := httpx.DecodeJSON(r, &request); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(request); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}

	if err := h.service.Logout(r.Context(), request.RefreshToken); err != nil {
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

	user, err := h.service.Me(r.Context(), principal.UserID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, users.MeResponse{User: user})
}

func sessionMetaFromRequest(r *http.Request) sessions.SessionMeta {
	userAgent := r.UserAgent()
	ipAddress := r.RemoteAddr
	return sessions.SessionMeta{
		UserAgent: &userAgent,
		IPAddress: &ipAddress,
	}
}
