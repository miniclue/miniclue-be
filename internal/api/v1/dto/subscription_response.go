package dto

import "time"

// SubscriptionResponseDTO represents the authenticated user's current subscription.
type SubscriptionResponseDTO struct {
	PlanID               string    `json:"plan_id"`
	Name                 string    `json:"name"`
	StripeSubscriptionID *string   `json:"stripe_subscription_id,omitempty"`
	StartsAt             time.Time `json:"starts_at"`
	EndsAt               time.Time `json:"ends_at"`
	Status               string    `json:"status"`
}
