package middleware

import (
	"context"
	"net/http"
	"strings"
	// "app/internal/errors"
	"app/internal/util" // JWT helper
)

// Injected key type to avoid context collisions
type contextKey string

const UserContextKey = contextKey("user")

func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "authorization header missing", http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}
			tokenString := parts[1]
			claims, err := util.ValidateJWT(tokenString, jwtSecret)
			if err != nil {
				http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
				return
			}
			// Embed user ID (or entire claims) into request context
			ctx := context.WithValue(r.Context(), UserContextKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
