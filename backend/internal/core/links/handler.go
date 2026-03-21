package links

import (
	"net/http"

	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/httpx"

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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}

	var request CreateLinkRequest
	if err := httpx.DecodeJSON(r, &request); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(request); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}

	link, err := h.service.Create(r.Context(), principal.UserID, request)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, link)
}

func (h *Handler) ByEntity(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}

	entityID, err := uuid.Parse(r.URL.Query().Get("entity_id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_entity_id", "entity_id must be a valid uuid"))
		return
	}

	query := ListByEntityQuery{
		EntityType: r.URL.Query().Get("entity_type"),
		EntityID:   entityID,
	}
	if err := h.validator.Struct(query); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}

	result, err := h.service.ListByEntity(r.Context(), principal.UserID, query)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": result})
}
