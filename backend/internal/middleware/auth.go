package middleware

import (
	"net/http"
	"strings"

	platformauth "moneyapp/backend/internal/platform/auth"
	"moneyapp/backend/internal/platform/httpx"
)

func AuthRequired(jwt *platformauth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				httpx.WriteError(w, httpx.Unauthorized("missing_bearer_token", "authorization token is required"))
				return
			}

			principal, err := jwt.ParseAccessToken(strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				httpx.WriteError(w, httpx.Unauthorized("invalid_token", "authorization token is invalid"))
				return
			}

			ctx := platformauth.ContextWithPrincipal(r.Context(), *principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RBAC(permissionCode string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := platformauth.PrincipalFromContext(r.Context())
			if !ok {
				httpx.WriteError(w, httpx.Unauthorized("unauthorized", "authorization required"))
				return
			}

			if !principal.HasPermission(permissionCode) {
				httpx.WriteError(w, httpx.Forbidden("forbidden", "permission denied"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func DepartmentScope() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}
