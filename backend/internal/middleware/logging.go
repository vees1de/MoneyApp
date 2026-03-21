package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/httpx"
)

func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorder := httpx.NewStatusRecorder(w)
			startedAt := time.Now()

			next.ServeHTTP(recorder, r)

			attrs := []any{
				"request_id", RequestIDFromContext(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"status", recorder.Status(),
				"duration", time.Since(startedAt).String(),
				"remote_addr", r.RemoteAddr,
			}

			if principal, ok := auth.PrincipalFromContext(r.Context()); ok {
				attrs = append(attrs, "user_id", principal.UserID.String())
			}

			logger.Info("http request", attrs...)
		})
	}
}
