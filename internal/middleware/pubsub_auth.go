package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
	"google.golang.org/api/idtoken"
)

// PubSubAuthMiddleware validates a JWT from a Pub/Sub push request.
// It bypasses authentication if isLocalDev is true.
func PubSubAuthMiddleware(isLocalDev bool, audience, expectedEmail string, logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For local development, bypass the authentication check.
			if isLocalDev {
				logger.Debug().Msg("Skipping Pub/Sub authentication for local environment")
				next.ServeHTTP(w, r)
				return
			}

			if audience == "" || expectedEmail == "" {
				logger.Error().Msg("Pub/Sub auth middleware configured without an audience or expected email; requests will be denied")
				http.Error(w, "Configuration error: audience or email not set", http.StatusInternalServerError)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Warn().Msg("Missing Authorization header in Pub/Sub push request")
				http.Error(w, "Unauthorized: missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				logger.Warn().Msg("Malformed Authorization header in Pub/Sub push request")
				http.Error(w, "Unauthorized: malformed authorization header", http.StatusUnauthorized)
				return
			}
			tokenString := parts[1]

			payload, err := idtoken.Validate(context.Background(), tokenString, audience)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to validate Pub/Sub JWT")
				http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
				return
			}

			email, ok := payload.Claims["email"].(string)
			if !ok || email == "" {
				logger.Error().Err(err).Msg("Email claim missing or invalid in Pub/Sub JWT")
				http.Error(w, "Forbidden: invalid email claim in token", http.StatusForbidden)
				return
			}

			if email != expectedEmail {
				logger.Warn().
					Str("token_email", email).
					Str("expected_email", expectedEmail).
					Msg("Pub/Sub JWT email does not match expected service account")
				http.Error(w, "Forbidden: token email does not match expected service account", http.StatusForbidden)
				return
			}

			logger.Info().
				Str("email", email).
				Str("issuer", payload.Issuer).
				Msg("Successfully authenticated and authorized Pub/Sub push request")

			next.ServeHTTP(w, r)
		})
	}
}
