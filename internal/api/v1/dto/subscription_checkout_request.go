package dto

// SubscriptionCheckoutRequest represents a request to initiate a Stripe checkout session.
type SubscriptionCheckoutRequest struct {
	Plan string `json:"plan" validate:"required,oneof=monthly annual monthly_launch annual_launch"`
}
