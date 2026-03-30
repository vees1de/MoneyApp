package cicd

import (
	"net/http"

	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Run(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}

	result, err := h.service.Run(r.Context(), principal)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, result)
}
