package middleware

import "net/http"

var allowedLocalOrigins = map[string]struct{}{
	"http://localhost:4200": {},
	"http://localhost:8080": {},
}

func CORSLocalhost4200(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if _, ok := allowedLocalOrigins[origin]; ok {
			headers := w.Header()
			headers.Set("Access-Control-Allow-Origin", origin)
			headers.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
			headers.Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			headers.Set("Access-Control-Expose-Headers", "X-Request-ID")
			headers.Set("Access-Control-Max-Age", "600")
			headers.Set("Vary", "Origin")
		}

		if _, ok := allowedLocalOrigins[origin]; ok && r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
