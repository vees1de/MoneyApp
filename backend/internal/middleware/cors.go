package middleware

import "net/http"

const localhost4200Origin = "http://localhost:4200"

func CORSLocalhost4200(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == localhost4200Origin {
			headers := w.Header()
			headers.Set("Access-Control-Allow-Origin", origin)
			headers.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
			headers.Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			headers.Set("Access-Control-Expose-Headers", "X-Request-ID")
			headers.Set("Access-Control-Max-Age", "600")
			headers.Set("Vary", "Origin")
		}

		if r.Method == http.MethodOptions && origin == localhost4200Origin {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
