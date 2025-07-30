package model

import "time"

// SubscriptionPlan represents a pricing plan's configuration and limits.
type SubscriptionPlan struct {
	ID            string                 `db:"id"`
	Name          string                 `db:"name"`
	PriceCents    int                    `db:"price_cents"`
	BillingPeriod string                 `db:"billing_period"` // Postgres interval as string
	MaxUploads    int                    `db:"max_uploads"`
	MaxSizeMB     int                    `db:"max_size_mb"`
	ChatLimit     int                    `db:"chat_limit"`
	FeatureFlags  map[string]interface{} `db:"feature_flags"`
}

// UserSubscription represents an individual user's subscription record.
type UserSubscription struct {
	UserID               string    `db:"user_id"`
	PlanID               string    `db:"plan_id"`
	StripeSubscriptionID string    `db:"stripe_subscription_id" json:"stripe_subscription_id,omitempty"`
	StartsAt             time.Time `db:"starts_at"`
	EndsAt               time.Time `db:"ends_at"`
	Status               string    `db:"status"`
}
