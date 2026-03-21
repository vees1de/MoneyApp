package review

import (
	"net/http"

	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/httpx"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
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

func (h *Handler) Current(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}
	review, err := h.service.GetCurrent(r.Context(), principal.UserID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, review)
}

func (h *Handler) SubmitBalance(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}
	reviewID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_review_id", "invalid review id"))
		return
	}
	var request SubmitBalanceRequest
	if err := httpx.DecodeJSON(r, &request); err != nil {
		httpx.WriteError(w, err)
		return
	}
	review, err := h.service.SubmitBalance(r.Context(), principal.UserID, reviewID, request.ActualBalance)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, review)
}

func (h *Handler) Resolve(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}
	reviewID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_review_id", "invalid review id"))
		return
	}
	var request ResolveRequest
	if err := httpx.DecodeJSON(r, &request); err != nil {
		httpx.WriteError(w, err)
		return
	}
	review, err := h.service.Resolve(r.Context(), principal.UserID, reviewID, request.ResolutionNote)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, review)
}

func (h *Handler) Skip(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}
	reviewID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_review_id", "invalid review id"))
		return
	}
	review, err := h.service.Skip(r.Context(), principal.UserID, reviewID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, review)
}
