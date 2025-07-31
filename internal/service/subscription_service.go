package service

import (
	"app/internal/model"
	"app/internal/repository"
	"context"
	"time"

	"github.com/rs/zerolog"
)

// SubscriptionService defines business logic methods for subscriptions.
type SubscriptionService interface {
	GetActiveSubscription(ctx context.Context, userID string) (*model.UserSubscription, error)
	GetPlan(ctx context.Context, planID string) (*model.SubscriptionPlan, error)
	UpsertStripeSubscription(ctx context.Context, userID, planID string, startsAt, endsAt time.Time, status, stripeSubscriptionID string) error
	DowngradeUserToFreePlan(ctx context.Context, userID, freePlanID string) error
}

type subscriptionService struct {
	repo   repository.SubscriptionRepository
	logger zerolog.Logger
}

// NewSubscriptionService creates a new SubscriptionService with a scoped logger.
func NewSubscriptionService(repo repository.SubscriptionRepository, logger zerolog.Logger) SubscriptionService {
	lg := logger.With().Str("service", "SubscriptionService").Logger()
	return &subscriptionService{repo: repo, logger: lg}
}

// GetActiveSubscription returns the current active subscription for a user.
func (s *subscriptionService) GetActiveSubscription(ctx context.Context, userID string) (*model.UserSubscription, error) {
	sub, err := s.repo.GetActiveSubscription(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to fetch active subscription")
		return nil, err
	}

	return sub, nil
}

// GetPlan returns the details of a subscription plan.
func (s *subscriptionService) GetPlan(ctx context.Context, planID string) (*model.SubscriptionPlan, error) {
	plan, err := s.repo.GetPlanByID(ctx, planID)
	if err != nil {
		s.logger.Error().Err(err).Str("plan_id", planID).Msg("Failed to fetch subscription plan")
	}
	return plan, err
}

func (s *subscriptionService) UpsertStripeSubscription(ctx context.Context, userID, planID string, startsAt, endsAt time.Time, status, stripeSubscriptionID string) error {
	if err := s.repo.UpsertStripeSubscription(ctx, userID, planID, startsAt, endsAt, status, stripeSubscriptionID); err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Str("plan_id", planID).Str("status", status).Msg("Failed to upsert stripe subscription")
		return err
	}
	return nil
}

// DowngradeUserToFreePlan downgrades a user to the free plan when their subscription is deleted
func (s *subscriptionService) DowngradeUserToFreePlan(ctx context.Context, userID, freePlanID string) error {
	if err := s.repo.DowngradeUserToFreePlan(ctx, userID, freePlanID); err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to downgrade user to free plan")
		return err
	}
	return nil
}
