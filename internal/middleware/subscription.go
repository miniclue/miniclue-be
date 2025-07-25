package middleware

import (
	"context"
	"errors"
	"net/http"

	"app/internal/service"

	"github.com/jackc/pgx/v5"
)

// subscriptionContextKey is key type for storing values in context
type subscriptionContextKey string

// PlanContextKey stores the user's SubscriptionPlan in the request context
const PlanContextKey subscriptionContextKey = "subscriptionPlan"

// SubscriptionLimitMiddleware checks upload limits for lecture uploads.
func SubscriptionLimitMiddleware(subSvc service.SubscriptionService, lecSvc service.LectureService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only enforce on POST /lectures
			if r.Method == http.MethodPost && r.URL.Path == "/lectures" {
				// Get user ID from context
				userID, ok := r.Context().Value(UserContextKey).(string)
				if !ok || userID == "" {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				// Fetch subscription
				sub, err := subSvc.GetActiveSubscription(r.Context(), userID)
				if err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						http.Error(w, "No active subscription found", http.StatusForbidden)
						return
					}
					http.Error(w, "Failed to fetch subscription: "+err.Error(), http.StatusInternalServerError)
					return
				}
				// Fetch plan details
				plan, err := subSvc.GetPlan(r.Context(), sub.PlanID)
				if err != nil {
					http.Error(w, "Failed to fetch plan: "+err.Error(), http.StatusInternalServerError)
					return
				}
				// Inject plan into context for downstream handlers
				ctx := context.WithValue(r.Context(), PlanContextKey, plan)
				r = r.WithContext(ctx)
				// Enforce upload limit only if finite (>0)
				if plan.MaxUploads > 0 {
					// Count lectures in current period
					count, err := lecSvc.CountLecturesByUser(r.Context(), userID, sub.StartsAt, sub.EndsAt)
					if err != nil {
						http.Error(w, "Failed to count uploads: "+err.Error(), http.StatusInternalServerError)
						return
					}
					if count >= plan.MaxUploads {
						http.Error(w, "Upload limit reached", http.StatusTooManyRequests)
						return
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
