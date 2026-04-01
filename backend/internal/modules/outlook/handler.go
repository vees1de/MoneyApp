package outlook

import (
	"net/http"
	"strconv"

	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func outlookPrincipal(r *http.Request) (platformauth.Principal, error) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		return platformauth.Principal{}, httpx.Unauthorized("unauthorized", "authorization required")
	}
	return principal, nil
}

func (h *Handler) Connect(w http.ResponseWriter, r *http.Request) {
	principal, err := outlookPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	response, err := h.service.Connect(principal)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	target, err := h.service.Callback(
		r.Context(),
		r.URL.Query().Get("state"),
		r.URL.Query().Get("code"),
		r.URL.Query().Get("error"),
		r.URL.Query().Get("error_description"),
	)
	if err != nil {
		http.Redirect(w, r, target, http.StatusFound)
		return
	}

	http.Redirect(w, r, target, http.StatusFound)
}

func (h *Handler) ManualConnect(w http.ResponseWriter, r *http.Request) {
	principal, err := outlookPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	var req ManualConnectRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	status, err := h.service.ManualConnect(r.Context(), principal, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, status)
}

func (h *Handler) Disconnect(w http.ResponseWriter, r *http.Request) {
	principal, err := outlookPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	if err := h.service.Disconnect(r.Context(), principal); err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteNoContent(w)
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	principal, err := outlookPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	status, err := h.service.Status(r.Context(), principal)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, status)
}

func (h *Handler) Sync(w http.ResponseWriter, r *http.Request) {
	principal, err := outlookPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	response, err := h.service.Sync(r.Context(), principal)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	principal, err := outlookPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	limit := defaultListLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	items, err := h.service.ListMessages(r.Context(), principal, limit)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	principal, err := outlookPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	limit := defaultListLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	items, err := h.service.ListEvents(r.Context(), principal, limit)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	principal, err := outlookPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	var req UpdateSettingsRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	status, err := h.service.UpdateSettings(r.Context(), principal, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, status)
}

func (h *Handler) SendTestEmail(w http.ResponseWriter, r *http.Request) {
	principal, err := outlookPrincipal(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	var req TestEmailRequest
	if r.ContentLength > 0 {
		if err := httpx.DecodeJSON(r, &req); err != nil {
			httpx.WriteError(w, err)
			return
		}
	}

	response, err := h.service.SendTestEmail(r.Context(), principal, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}
