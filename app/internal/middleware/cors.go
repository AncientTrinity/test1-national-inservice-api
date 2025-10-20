package middleware

import (
	"net/http"
	"strings"
)

// CORS returns middleware that sets CORS headers based on allowed origins list.
func CORS(allowedOrigins []string) func(next http.Handler) http.Handler {
	// Normalize allowed origins (allow "*")
	allowAll := false
	for _, o := range allowedOrigins {
		if o == "*" {
			allowAll = true
			break
		}
	}
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origin != "" {
				for _, o := range allowedOrigins {
					if o == origin {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						break
					}
					// allow simple wildcard match like https://*.example.com not implemented
				}
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
