package transactions

import (
	"net/http"
	"time"

	"moneyapp/backend/internal/core/common"
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

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}

	filters, err := filtersFromRequest(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	items, err := h.service.List(r.Context(), principal.UserID, filters)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}
	var request CreateTransactionRequest
	if err := httpx.DecodeJSON(r, &request); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.validator.Struct(request); err != nil {
		httpx.WriteError(w, httpx.BadRequest("validation_error", err.Error()))
		return
	}

	transaction, err := h.service.Create(r.Context(), principal.UserID, request)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, transaction)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}
	transactionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_transaction_id", "invalid transaction id"))
		return
	}
	transaction, err := h.service.Get(r.Context(), principal.UserID, transactionID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, transaction)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}
	transactionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_transaction_id", "invalid transaction id"))
		return
	}
	var request UpdateTransactionRequest
	if err := httpx.DecodeJSON(r, &request); err != nil {
		httpx.WriteError(w, err)
		return
	}

	transaction, err := h.service.Update(r.Context(), principal.UserID, transactionID, request)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, transaction)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
		return
	}
	transactionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, httpx.BadRequest("invalid_transaction_id", "invalid transaction id"))
		return
	}
	if err := h.service.Delete(r.Context(), principal.UserID, transactionID); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteNoContent(w)
}

func filtersFromRequest(r *http.Request) (ListFilters, error) {
	filters := ListFilters{
		Pagination: common.PaginationFromRequest(r, 50),
	}
	query := r.URL.Query()

	if raw := query.Get("account"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return ListFilters{}, httpx.BadRequest("invalid_account", "account filter must be a valid uuid")
		}
		filters.AccountID = &id
	}
	if raw := query.Get("category"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return ListFilters{}, httpx.BadRequest("invalid_category", "category filter must be a valid uuid")
		}
		filters.CategoryID = &id
	}
	if raw := query.Get("type"); raw != "" {
		txType := Type(raw)
		filters.Type = &txType
	}
	if raw := query.Get("linked_entity_type"); raw != "" {
		filters.LinkedEntityType = &raw
	}
	if raw := query.Get("linked_entity_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return ListFilters{}, httpx.BadRequest("invalid_linked_entity_id", "linked_entity_id must be a valid uuid")
		}
		filters.LinkedEntityID = &id
	}
	if raw := query.Get("date_from"); raw != "" {
		value, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return ListFilters{}, httpx.BadRequest("invalid_date_from", "date_from must be RFC3339")
		}
		filters.DateFrom = &value
	}
	if raw := query.Get("date_to"); raw != "" {
		value, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return ListFilters{}, httpx.BadRequest("invalid_date_to", "date_to must be RFC3339")
		}
		filters.DateTo = &value
	}

	return filters, nil
}
