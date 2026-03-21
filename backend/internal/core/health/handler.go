package health

import (
	"context"
	"net/http"
	"time"

	"moneyapp/backend/internal/platform/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Live(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	dependencies := map[string]string{}
	hasErrors := false
	for name, check := range h.service.Checks() {
		if err := check(ctx); err != nil {
			dependencies[name] = err.Error()
			hasErrors = true
			continue
		}
		dependencies[name] = "ok"
	}

	statusCode := http.StatusOK
	status := "ok"
	if hasErrors {
		statusCode = http.StatusServiceUnavailable
		status = "degraded"
	}

	httpx.WriteJSON(w, statusCode, map[string]any{
		"status":       status,
		"dependencies": dependencies,
	})
}
