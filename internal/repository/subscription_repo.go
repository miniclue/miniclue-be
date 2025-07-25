package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"app/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SubscriptionRepository defines methods for accessing subscription data.
type SubscriptionRepository interface {
	GetActiveSubscription(ctx context.Context, userID string) (*model.UserSubscription, error)
	GetPlanByID(ctx context.Context, planID string) (*model.SubscriptionPlan, error)
	// UpsertSubscription creates a subscription with the given planId for a new user if none exists, using the plan's billing_period.
	UpsertSubscription(ctx context.Context, userID, planID string) error
}

type subscriptionRepo struct {
	pool *pgxpool.Pool
}

// NewSubscriptionRepo creates a new SubscriptionRepository.
func NewSubscriptionRepo(pool *pgxpool.Pool) SubscriptionRepository {
	return &subscriptionRepo{pool: pool}
}

// GetActiveSubscription returns the current active subscription for a user.
func (r *subscriptionRepo) GetActiveSubscription(ctx context.Context, userID string) (*model.UserSubscription, error) {
	const q = `
        SELECT user_id, plan_id, starts_at, ends_at, status
        FROM user_subscriptions
        WHERE user_id = $1
          AND status IN ('active', 'cancelled') -- Allow paid users to use service until period end
          AND (ends_at + INTERVAL '6 hours') > NOW() -- 6h grace period covers the gap before the daily cron job renews free/beta plans
    `
	var us model.UserSubscription
	err := r.pool.QueryRow(ctx, q, userID).Scan(
		&us.UserID,
		&us.PlanID,
		&us.StartsAt,
		&us.EndsAt,
		&us.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("fetch active subscription for user %s: %w", userID, err)
	}
	return &us, nil
}

// GetPlanByID returns the subscription plan with its limits.
func (r *subscriptionRepo) GetPlanByID(ctx context.Context, planID string) (*model.SubscriptionPlan, error) {
	const q = `
        SELECT id,
               name,
               price_cents,
               billing_period::text AS billing_period,
               max_uploads,
               max_size_mb,
               chat_limit,
               feature_flags
        FROM subscription_plans
        WHERE id = $1
    `
	var sp model.SubscriptionPlan
	var rawFlags []byte
	err := r.pool.QueryRow(ctx, q, planID).Scan(
		&sp.ID,
		&sp.Name,
		&sp.PriceCents,
		&sp.BillingPeriod,
		&sp.MaxUploads,
		&sp.MaxSizeMB,
		&sp.ChatLimit,
		&rawFlags,
	)
	if err != nil {
		return nil, fmt.Errorf("fetch plan %s: %w", planID, err)
	}
	if err := json.Unmarshal(rawFlags, &sp.FeatureFlags); err != nil {
		return nil, fmt.Errorf("unmarshal feature_flags for plan %s: %w", planID, err)
	}
	return &sp, nil
}

// UpsertSubscription creates a subscription for a user with the given planID if none exists.
func (r *subscriptionRepo) UpsertSubscription(ctx context.Context, userID, planID string) error {
	const q = `
        INSERT INTO user_subscriptions (user_id, plan_id, starts_at, ends_at, status)
        SELECT $1, $2, NOW(), NOW() + billing_period, 'active'
        FROM subscription_plans
        WHERE id = $2
        ON CONFLICT (user_id) DO NOTHING;
    `
	_, err := r.pool.Exec(ctx, q, userID, planID)
	if err != nil {
		return fmt.Errorf("upserting subscription %s for user %s: %w", planID, userID, err)
	}
	return nil
}
