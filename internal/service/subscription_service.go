package service

import (
	"app/internal/model"
	"app/internal/repository"
	"context"
)

// SubscriptionService defines business logic methods for subscriptions.
type SubscriptionService interface {
	GetActiveSubscription(ctx context.Context, userID string) (*model.UserSubscription, error)
	GetPlan(ctx context.Context, planID string) (*model.SubscriptionPlan, error)
}

type subscriptionService struct {
	repo repository.SubscriptionRepository
}

// NewSubscriptionService creates a new SubscriptionService.
func NewSubscriptionService(repo repository.SubscriptionRepository) SubscriptionService {
	return &subscriptionService{repo: repo}
}

// GetActiveSubscription returns the current active subscription for a user.
func (s *subscriptionService) GetActiveSubscription(ctx context.Context, userID string) (*model.UserSubscription, error) {
	return s.repo.GetActiveSubscription(ctx, userID)
}

// GetPlan returns the details of a subscription plan.
func (s *subscriptionService) GetPlan(ctx context.Context, planID string) (*model.SubscriptionPlan, error) {
	return s.repo.GetPlanByID(ctx, planID)
}
