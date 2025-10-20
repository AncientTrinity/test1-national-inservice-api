package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const (
	CTXAccountID ctxKey = "account_id"
	CTXRole      ctxKey = "role"
)

// RequireAuth parses Authorization header "Bearer <token>" and sets context values
func RequireAuth(jwtKey []byte) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				http.Error(w, "missing authorization", http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "invalid authorization", http.StatusUnauthorized)
				return
			}
			tknStr := parts[1]
			tkn, err := jwt.Parse(tknStr, func(t *jwt.Token) (interface{}, error) {
				if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
					return nil, jwt.NewValidationError("unexpected signing method", jwt.ValidationErrorSignatureInvalid)
				}
				return jwtKey, nil
			})
			if err != nil || !tkn.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			claims, ok := tkn.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "invalid token claims", http.StatusUnauthorized)
				return
			}
			// extract sub and role
			var accID interface{}
			if v, ok := claims["sub"]; ok {
				accID = v
			}
			var role interface{}
			if v, ok := claims["role"]; ok {
				role = v
			}

			ctx := context.WithValue(r.Context(), CTXAccountID, accID)
			ctx = context.WithValue(ctx, CTXRole, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

// RequireRole ensures the account has the specific role
func RequireRole(roleName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			role := r.Context().Value(CTXRole)
			if role == nil {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			rs, ok := role.(string)
			if !ok || rs != roleName {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
