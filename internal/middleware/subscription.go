package middleware

import (
	"context"
	"errors"
	"net/http"

	"app/internal/service"

	"github.com/jackc/pgx/v5"
)

// contextKey type for subscription middleware
type subscriptionContextKey string

// PlanContextKey stores the SubscriptionPlan in context
const PlanContextKey subscriptionContextKey = "subscriptionPlan"

// SubscriptionContextKey stores the UserSubscription in context
const SubscriptionContextKey subscriptionContextKey = "userSubscription"

// SubscriptionLimitMiddleware fetches and injects subscription data for lecture uploads.
func SubscriptionLimitMiddleware(subSvc service.SubscriptionService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only prepare plan context for POST /lectures and POST /lectures/batch-upload-url
			if r.Method == http.MethodPost && (r.URL.Path == "/lectures" || r.URL.Path == "/lectures/batch-upload-url") {
				userID, ok := r.Context().Value(UserContextKey).(string)
				if !ok || userID == "" {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				// Fetch subscription, block if none
				sub, err := subSvc.GetActiveSubscription(r.Context(), userID)
				if err != nil {
					// No active sub => forbidden
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
				// Inject subscription and plan into context for downstream handlers
				ctx := context.WithValue(r.Context(), PlanContextKey, plan)
				ctx = context.WithValue(ctx, SubscriptionContextKey, sub)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}
