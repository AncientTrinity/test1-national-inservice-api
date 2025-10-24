package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// context key for user claims
type ctxKey string

const ctxUserKey ctxKey = "userClaims"

// RequireAuth verifies JWT in Authorization header and puts claims into context.
func RequireAuth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				http.Error(w, "missing Authorization header", http.StatusUnauthorized)
				return
			}
			tokenStr := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer"))

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				return secret, nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ctxUserKey, token.Claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole ensures the user's "role" claim matches required role.
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			val := r.Context().Value(ctxUserKey)
			if val == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			claims, ok := val.(jwt.Claims)
			if !ok {
				http.Error(w, "invalid token claims", http.StatusUnauthorized)
				return
			}
			// support MapClaims (common case)
			if mc, ok := claims.(jwt.MapClaims); ok {
				if mcRole, _ := mc["role"].(string); mcRole == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, "forbidden", http.StatusForbidden)
		})
	}
}
