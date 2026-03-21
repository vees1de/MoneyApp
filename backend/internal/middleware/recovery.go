package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"moneyapp/backend/internal/platform/httpx"
)

func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error("panic recovered",
						"request_id", RequestIDFromContext(r.Context()),
						"error", fmt.Sprintf("%v", recovered),
					)
					httpx.WriteError(w, httpx.Internal("panic_recovered"))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
