package middleware

import (
	"app/internal/logger"
	"net/http"
)

// LoggerMiddleware logs incoming HTTP requests.
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Call the next handler in the chain
		next.ServeHTTP(w, r)

		logger := logger.New()
		// Log original message format with full request URI including query params
		logger.Debug().Msgf("%s %s", r.Method, r.URL.RequestURI())
	})
}
