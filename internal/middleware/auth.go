package middleware

import (
	"app/internal/logger"
	"app/internal/util" // JWT helper
	"context"
	"net/http"
	"strings"
)

// Injected key type to avoid context collisions
type contextKey string

const UserContextKey = contextKey("user")

func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logger.New()
			logger.Info().Msg("Auth middleware running")
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Error().Msg("Authorization header missing")
				http.Error(w, "Authorization header missing", http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				logger.Error().Msg("Invalid authorization header")
				http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
				return
			}
			tokenString := parts[1]
			claims, err := util.ValidateJWT(tokenString, jwtSecret)
			if err != nil {
				logger.Error().Msgf("Invalid token: %+v", err)
				http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
				return
			}
			// Embed user ID (or entire claims) into request context
			ctx := context.WithValue(r.Context(), UserContextKey, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
