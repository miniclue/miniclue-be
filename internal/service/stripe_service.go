package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"app/internal/config"
	"app/internal/model"
	"app/internal/repository"

	"github.com/rs/zerolog"
	"github.com/stripe/stripe-go/v82"
	billingsession "github.com/stripe/stripe-go/v82/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v82/checkout/session"
	customerpkg "github.com/stripe/stripe-go/v82/customer"
	subscriptionpkg "github.com/stripe/stripe-go/v82/subscription"
	"github.com/stripe/stripe-go/v82/webhook"
)

// StripeService manages Stripe integration
type StripeService struct {
	cfg      *config.Config
	userRepo repository.UserRepository
	subSvc   SubscriptionService
	logger   zerolog.Logger
}

// NewStripeService initializes Stripe key and returns service with a scoped logger
func NewStripeService(cfg *config.Config, userRepo repository.UserRepository, subSvc SubscriptionService, logger zerolog.Logger) *StripeService {
	stripe.Key = cfg.StripeSecretKey
	lg := logger.With().Str("service", "StripeService").Logger()
	return &StripeService{cfg: cfg, userRepo: userRepo, subSvc: subSvc, logger: lg}
}

// getUserIDFromEvent is a helper method to resolve user ID from webhook metadata or customer ID
func (s *StripeService) getUserIDFromEvent(ctx context.Context, metadata map[string]string, customerID string) (string, error) {
	if userID, ok := metadata["user_id"]; ok && userID != "" {
		return userID, nil
	}
	if customerID == "" {
		return "", errors.New("cannot determine user: missing metadata and customer id")
	}
	s.logger.Warn().Str("stripe_customer_id", customerID).Msg("Missing user_id metadata; looking up user by customer ID")
	u, err := s.userRepo.GetUserByStripeCustomerID(ctx, customerID)
	if err != nil {
		return "", fmt.Errorf("failed to lookup user by Stripe customer ID: %w", err)
	}
	if u == nil {
		return "", fmt.Errorf("no user found for customer ID: %s", customerID)
	}
	return u.UserID, nil
}

// GetOrCreateCustomer ensures a Stripe Customer exists for a user
// Since customers are now created at signup, this method primarily handles edge cases
func (s *StripeService) GetOrCreateCustomer(ctx context.Context, user *model.User) (string, error) {
	if user.StripeCustomerID != nil && *user.StripeCustomerID != "" {
		return *user.StripeCustomerID, nil
	}

	// Fallback: create customer if it doesn't exist (for legacy users or edge cases)
	s.logger.Warn().Str("user_id", user.UserID).Msg("No Stripe customer ID found, creating customer as fallback")
	return s.CreateCustomer(ctx, user)
}

// CreateCustomer creates a new Stripe customer for a user
func (s *StripeService) CreateCustomer(ctx context.Context, user *model.User) (string, error) {
	params := &stripe.CustomerParams{
		Email:    stripe.String(user.Email),
		Name:     stripe.String(user.Name),
		Metadata: map[string]string{"user_id": user.UserID},
	}
	cust, err := customerpkg.New(params)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", user.UserID).Msg("Failed to create Stripe customer")
		return "", fmt.Errorf("create stripe customer: %w", err)
	}
	if err := s.userRepo.UpdateStripeCustomerID(ctx, user.UserID, cust.ID); err != nil {
		s.logger.Error().Err(err).Str("user_id", user.UserID).Msg("Failed to store stripe customer id in user_profiles")
		return "", fmt.Errorf("store stripe customer id: %w", err)
	}
	return cust.ID, nil
}

// CreateCheckoutSession creates a Stripe Checkout session
func (s *StripeService) CreateCheckoutSession(ctx context.Context, userID, plan string) (string, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to fetch user for checkout session")
		return "", fmt.Errorf("fetch user: %w", err)
	}
	if user == nil {
		s.logger.Error().Str("user_id", userID).Msg("User not found for checkout session")
		return "", fmt.Errorf("user not found: %s", userID)
	}
	customerID, err := s.GetOrCreateCustomer(ctx, user)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get or create Stripe customer for checkout session")
		return "", err
	}
	var priceID string
	switch plan {
	case "monthly":
		priceID = s.cfg.StripePriceMonthly
	case "annual":
		priceID = s.cfg.StripePriceAnnual
	case "monthly_launch":
		priceID = s.cfg.StripePriceMonthlyLaunch
	case "annual_launch":
		priceID = s.cfg.StripePriceAnnualLaunch
	default:
		return "", fmt.Errorf("invalid plan: %s", plan)
	}
	sessParams := &stripe.CheckoutSessionParams{
		Customer:           stripe.String(customerID),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems:          []*stripe.CheckoutSessionLineItemParams{{Price: stripe.String(priceID), Quantity: stripe.Int64(1)}},
		Mode:               stripe.String(stripe.CheckoutSessionModeSubscription),
		SuccessURL:         stripe.String(s.cfg.StripePortalReturnURL),
		CancelURL:          stripe.String(s.cfg.StripePortalReturnURL),
		Metadata:           map[string]string{"user_id": userID},
	}
	sess, err := checkoutsession.New(sessParams)
	if err != nil {
		s.logger.Error().Err(err).Str("plan", plan).Msg("Failed to create Stripe checkout session")
		return "", fmt.Errorf("create checkout session: %w", err)
	}
	return sess.URL, nil
}

// CreatePortalSession creates a Stripe Customer Portal session
func (s *StripeService) CreatePortalSession(ctx context.Context, userID string) (string, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to fetch user for portal session")
		return "", fmt.Errorf("fetch user: %w", err)
	}
	if user == nil || user.StripeCustomerID == nil || *user.StripeCustomerID == "" {
		s.logger.Error().Str("user_id", userID).Msg("No Stripe customer ID found for user when creating portal session")
		return "", fmt.Errorf("no stripe customer for user: %s", userID)
	}
	params := &stripe.BillingPortalSessionParams{Customer: stripe.String(*user.StripeCustomerID), ReturnURL: stripe.String(s.cfg.StripePortalReturnURL)}
	sess, err := billingsession.New(params)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to create Stripe billing portal session")
		return "", fmt.Errorf("create billing portal session: %w", err)
	}
	return sess.URL, nil
}

// HandleWebhook processes Stripe webhook events
func (s *StripeService) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to read Stripe webhook payload")
		http.Error(w, "failed to read payload", http.StatusBadRequest)
		return
	}
	sig := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, sig, s.cfg.StripeWebhookSecret)
	if err != nil {
		s.logger.Error().Err(err).Msg("Signature verification failed for Stripe webhook")
		http.Error(w, "signature verification failed", http.StatusBadRequest)
		return
	}
	// Log receipt of webhook
	s.logger.Info().Str("event_type", string(event.Type)).Msg("Stripe webhook received")

	// Log the raw payload for debugging (be careful with sensitive data)
	s.logger.Debug().Str("event_type", string(event.Type)).RawJSON("payload", event.Data.Raw).Msg("Webhook payload received")

	ctx := r.Context()
	switch event.Type {
	case "checkout.session.completed":
		var cs stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &cs); err != nil {
			s.logger.Error().Err(err).Msg("Invalid checkout.session data")
			http.Error(w, "invalid checkout.session data", http.StatusBadRequest)
			return
		}
		subID := cs.Subscription.ID
		// Fetch full subscription object to get timing and price details
		subObj, err := subscriptionpkg.Get(subID, nil)
		if err != nil {
			s.logger.Error().Err(err).Str("subscription_id", subID).Msg("Failed to fetch subscription details")
			http.Error(w, "failed to fetch subscription details", http.StatusInternalServerError)
			return
		}
		// Extract plan ID
		planID := ""
		if len(subObj.Items.Data) > 0 && subObj.Items.Data[0].Price != nil {
			planID = subObj.Items.Data[0].Price.ID
		}
		if planID == "" {
			s.logger.Error().Str("subscription_id", subID).Msg("Could not determine price ID from subscription")
			http.Error(w, "could not determine price ID", http.StatusInternalServerError)
			return
		}
		// Determine subscription period from the first subscription item
		if len(subObj.Items.Data) == 0 {
			s.logger.Error().Str("subscription_id", subID).Msg("Subscription has no items")
			http.Error(w, "subscription has no items", http.StatusInternalServerError)
			return
		}
		item := subObj.Items.Data[0]
		start := time.Unix(item.CurrentPeriodStart, 0)
		end := time.Unix(item.CurrentPeriodEnd, 0)

		s.logger.Info().Str("subscription_id", cs.Subscription.ID).Str("plan_id", planID).Msg("Extracted plan ID from checkout session")

		userID := cs.Metadata["user_id"]
		if userID == "" {
			s.logger.Error().Str("subscription_id", cs.Subscription.ID).Msg("Missing user_id in checkout session metadata")
			http.Error(w, "missing user_id in metadata", http.StatusBadRequest)
			return
		}

		if err := s.subSvc.UpsertStripeSubscription(ctx, userID, planID, start, end, "active", subID); err != nil {
			s.logger.Error().Err(err).Msg("Failed to save subscription on checkout.session.completed")
			http.Error(w, "failed to save subscription", http.StatusInternalServerError)
			return
		}
	case "invoice.payment_succeeded":
		// Use official Stripe Invoice struct
		var invoice stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			s.logger.Error().Err(err).Msg("Invalid invoice.payment_succeeded payload")
			http.Error(w, "invalid invoice data", http.StatusBadRequest)
			return
		}

		// Determine user ID using helper method
		userID, err := s.getUserIDFromEvent(ctx, invoice.Metadata, invoice.Customer.ID)
		if err != nil {
			s.logger.Error().Err(err).Str("invoice_id", invoice.ID).Msg("Failed to determine user ID from invoice")
			http.Error(w, "failed to identify user", http.StatusInternalServerError)
			return
		}

		// Extract period
		start := time.Unix(invoice.PeriodStart, 0)
		end := time.Unix(invoice.PeriodEnd, 0)

		// Find subscription ID from line items
		var subID string
		if invoice.Lines != nil && len(invoice.Lines.Data) > 0 {
			for _, line := range invoice.Lines.Data {
				if line.Subscription != nil && line.Subscription.ID != "" {
					subID = line.Subscription.ID
					break
				}
			}
		}

		// Skip if no subscription ID (this is likely a one-time invoice)
		if subID == "" {
			s.logger.Info().Str("invoice_id", invoice.ID).Msg("Invoice has no subscription, skipping subscription update")
			w.WriteHeader(http.StatusOK)
			return
		}

		var priceID string

		// Get price ID from subscription directly
		sub, err := subscriptionpkg.Get(subID, nil)
		if err != nil {
			s.logger.Error().Err(err).Str("subscription_id", subID).Msg("Failed to fetch subscription for price ID")
			http.Error(w, "failed to fetch subscription details", http.StatusInternalServerError)
			return
		}

		if len(sub.Items.Data) > 0 && sub.Items.Data[0].Price != nil {
			priceID = sub.Items.Data[0].Price.ID
			s.logger.Info().Str("subscription_id", subID).Str("price_id", priceID).Msg("Extracted price ID from subscription")
		}

		if priceID == "" {
			s.logger.Error().Str("subscription_id", subID).Msg("Could not determine price ID from subscription")
			http.Error(w, "could not determine price ID", http.StatusInternalServerError)
			return
		}

		s.logger.Info().Str("subscription_id", subID).Str("plan_id", priceID).Str("user_id", userID).Msg("Extracted plan ID from invoice.payment_succeeded")

		if err := s.subSvc.UpsertStripeSubscription(ctx, userID, priceID, start, end, "active", subID); err != nil {
			s.logger.Error().Err(err).Str("user_id", userID).Str("plan_id", priceID).Msg("Failed to update subscription on invoice.payment_succeeded")
			http.Error(w, "failed to update subscription", http.StatusInternalServerError)
			return
		}
	case "customer.subscription.updated":
		var ss stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &ss); err != nil {
			s.logger.Error().Err(err).Msg("Invalid customer.subscription.updated payload")
			http.Error(w, "invalid subscription data", http.StatusBadRequest)
			return
		}
		// Determine status: mark as 'cancelled' if scheduled to cancel or already canceled
		status := ss.Status
		if ss.CancelAtPeriodEnd || ss.Status == stripe.SubscriptionStatusCanceled {
			status = "cancelled"
		}
		if len(ss.Items.Data) == 0 {
			s.logger.Error().Str("subscription_id", ss.ID).Msg("Subscription has no items")
			http.Error(w, "subscription has no items", http.StatusBadRequest)
			return
		}

		item := ss.Items.Data[0]
		start := time.Unix(item.CurrentPeriodStart, 0)
		end := time.Unix(item.CurrentPeriodEnd, 0)

		planID := item.Price.ID
		if planID == "" {
			s.logger.Error().Str("subscription_id", ss.ID).Msg("Could not determine plan ID from subscription")
			http.Error(w, "could not determine plan ID", http.StatusInternalServerError)
			return
		}

		// Use helper method to get user ID
		userID, err := s.getUserIDFromEvent(ctx, ss.Metadata, ss.Customer.ID)
		if err != nil {
			s.logger.Error().Err(err).Str("subscription_id", ss.ID).Msg("Failed to determine user ID from subscription")
			http.Error(w, "failed to identify user", http.StatusInternalServerError)
			return
		}

		s.logger.Info().Str("subscription_id", ss.ID).Str("plan_id", planID).Str("user_id", userID).Msg("Extracted plan ID from customer.subscription.updated")

		if err := s.subSvc.UpsertStripeSubscription(ctx, userID, planID, start, end, string(status), ss.ID); err != nil {
			s.logger.Error().Err(err).Str("user_id", userID).Str("plan_id", planID).Msg("Failed to update subscription on customer.subscription.updated")
			http.Error(w, "failed to update subscription", http.StatusInternalServerError)
			return
		}
	case "customer.subscription.deleted":
		var ss stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &ss); err != nil {
			s.logger.Error().Err(err).Msg("Invalid customer.subscription.deleted payload")
			http.Error(w, "invalid subscription data", http.StatusBadRequest)
			return
		}

		// Use helper method to get user ID
		userID, err := s.getUserIDFromEvent(ctx, ss.Metadata, ss.Customer.ID)
		if err != nil {
			s.logger.Error().Err(err).Str("subscription_id", ss.ID).Msg("Failed to determine user ID from subscription")
			http.Error(w, "failed to identify user", http.StatusInternalServerError)
			return
		}

		// Use the dedicated downgrade method instead of UpsertStripeSubscription
		if err := s.subSvc.DowngradeUserToFreePlan(ctx, userID, s.cfg.StripePriceFree); err != nil {
			s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to downgrade subscription on customer.subscription.deleted")
			http.Error(w, "failed to downgrade subscription", http.StatusInternalServerError)
			return
		}
	case "invoice.payment_failed":
		// Use official Stripe Invoice struct
		var invoice stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			s.logger.Error().Err(err).Msg("Invalid invoice.payment_failed payload")
			http.Error(w, "invalid invoice data", http.StatusBadRequest)
			return
		}

		// Determine user ID using helper method
		userID, err := s.getUserIDFromEvent(ctx, invoice.Metadata, invoice.Customer.ID)
		if err != nil {
			s.logger.Error().Err(err).Str("invoice_id", invoice.ID).Msg("Failed to determine user ID from invoice")
			http.Error(w, "failed to identify user", http.StatusInternalServerError)
			return
		}

		// Extract period
		start := time.Unix(invoice.PeriodStart, 0)
		end := time.Unix(invoice.PeriodEnd, 0)

		// Find subscription ID from line items
		var subID string
		if invoice.Lines != nil && len(invoice.Lines.Data) > 0 {
			for _, line := range invoice.Lines.Data {
				if line.Subscription != nil && line.Subscription.ID != "" {
					subID = line.Subscription.ID
					break
				}
			}
		}

		// Skip if no subscription ID (this is likely a one-time invoice)
		if subID == "" {
			s.logger.Info().Str("invoice_id", invoice.ID).Msg("Invoice has no subscription, skipping subscription update")
			w.WriteHeader(http.StatusOK)
			return
		}

		var priceID string

		// Get price ID from subscription directly
		sub, err := subscriptionpkg.Get(subID, nil)
		if err != nil {
			s.logger.Error().Err(err).Str("subscription_id", subID).Msg("Failed to fetch subscription for price ID")
			http.Error(w, "failed to fetch subscription details", http.StatusInternalServerError)
			return
		}

		if len(sub.Items.Data) > 0 && sub.Items.Data[0].Price != nil {
			priceID = sub.Items.Data[0].Price.ID
			s.logger.Info().Str("subscription_id", subID).Str("price_id", priceID).Msg("Extracted price ID from subscription")
		}

		if priceID == "" {
			s.logger.Error().Str("subscription_id", subID).Msg("Could not determine price ID from subscription")
			http.Error(w, "could not determine price ID", http.StatusInternalServerError)
			return
		}

		if err := s.subSvc.UpsertStripeSubscription(ctx, userID, priceID, start, end, "past_due", subID); err != nil {
			s.logger.Error().Err(err).Str("user_id", userID).Str("plan_id", priceID).Msg("Failed to mark subscription as past_due on invoice.payment_failed")
			http.Error(w, "failed to mark past_due", http.StatusInternalServerError)
			return
		}
	default:
		s.logger.Warn().Str("event_type", string(event.Type)).Msg("Unhandled Stripe webhook event")
	}
	w.WriteHeader(http.StatusOK)
}
