package middleware

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

// RequireAuth checks for a valid JWT in Authorization header
func RequireAuth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := r.Header.Get("Authorization")
			if tokenStr == "" {
				http.Error(w, "missing token", http.StatusUnauthorized)
				return
			}
			// Parse token
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				return secret, nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			// store claims in context
			ctx := context.WithValue(r.Context(), "user", token.Claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole ensures the user has a specific role in claims
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value("user").(jwt.MapClaims)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if claims["role"] != role {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
