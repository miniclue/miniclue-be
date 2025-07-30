package service

import (
	"context"
	"encoding/json"
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

// GetOrCreateCustomer ensures a Stripe Customer exists for a user
func (s *StripeService) GetOrCreateCustomer(ctx context.Context, user *model.User) (string, error) {
	if user.StripeCustomerID != "" {
		return user.StripeCustomerID, nil
	}
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
	if user == nil || user.StripeCustomerID == "" {
		s.logger.Error().Str("user_id", userID).Msg("No Stripe customer ID found for user when creating portal session")
		return "", fmt.Errorf("no stripe customer for user: %s", userID)
	}
	params := &stripe.BillingPortalSessionParams{Customer: stripe.String(user.StripeCustomerID), ReturnURL: stripe.String(s.cfg.StripePortalReturnURL)}
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
		if cs.Invoice == nil {
			s.logger.Error().Msg("Checkout.session completed missing invoice")
			http.Error(w, "checkout session missing invoice", http.StatusBadRequest)
			return
		}
		// Use invoice period for subscription timing
		start := time.Unix(cs.Invoice.PeriodStart, 0)
		end := time.Unix(cs.Invoice.PeriodEnd, 0)
		planID := cs.Invoice.Lines.Data[0].Pricing.PriceDetails.Price
		userID := cs.Metadata["user_id"]
		if err := s.subSvc.UpsertStripeSubscription(ctx, userID, planID, start, end, "active", subID); err != nil {
			s.logger.Error().Err(err).Msg("Failed to save subscription on checkout.session.completed")
			http.Error(w, "failed to save subscription", http.StatusInternalServerError)
			return
		}
	case "invoice.payment_succeeded":
		// Parse invoice payload for subscription details
		var payload struct {
			Customer     string            `json:"customer"`
			Subscription string            `json:"subscription"`
			PeriodStart  int64             `json:"period_start"`
			PeriodEnd    int64             `json:"period_end"`
			Metadata     map[string]string `json:"metadata"`
			Lines        struct {
				Data []struct {
					Price struct {
						ID string `json:"id"`
					} `json:"price"`
				} `json:"data"`
			} `json:"lines"`
		}
		if err := json.Unmarshal(event.Data.Raw, &payload); err != nil {
			s.logger.Error().Err(err).Msg("Invalid invoice.payment_succeeded payload")
			http.Error(w, "invalid invoice data", http.StatusBadRequest)
			return
		}
		// Determine user ID: metadata takes precedence
		userID := payload.Metadata["user_id"]
		if userID == "" {
			custID := payload.Customer
			s.logger.Warn().Str("stripe_customer_id", custID).Msg("Invoice missing user_id metadata; looking up user by customer ID")
			u, err := s.userRepo.GetUserByStripeCustomerID(ctx, custID)
			if err != nil {
				s.logger.Error().Err(err).Msg("Failed to lookup user by Stripe customer ID")
				http.Error(w, "failed to identify user", http.StatusInternalServerError)
				return
			}
			if u == nil {
				s.logger.Error().Str("stripe_customer_id", custID).Msg("No user found for customer ID")
				http.Error(w, "user not found", http.StatusBadRequest)
				return
			}
			userID = u.UserID
		}
		// Extract period and price ID
		start := time.Unix(payload.PeriodStart, 0)
		end := time.Unix(payload.PeriodEnd, 0)
		if len(payload.Lines.Data) == 0 {
			s.logger.Warn().Msg("Invoice has no line items")
			break
		}
		priceID := payload.Lines.Data[0].Price.ID
		subID := payload.Subscription
		if err := s.subSvc.UpsertStripeSubscription(ctx, userID, priceID, start, end, "active", subID); err != nil {
			s.logger.Error().Err(err).Msg("Failed to update subscription on invoice.payment_succeeded")
			http.Error(w, "failed to update subscription", http.StatusInternalServerError)
			return
		}
	case "customer.subscription.updated":
		// Parse subscription payload directly for updated period
		var payload struct {
			ID                 string `json:"id"`
			CurrentPeriodStart int64  `json:"current_period_start"`
			CurrentPeriodEnd   int64  `json:"current_period_end"`
			Items              struct {
				Data []struct {
					Plan struct {
						ID string `json:"id"`
					} `json:"plan"`
				} `json:"data"`
			} `json:"items"`
			Metadata          map[string]string `json:"metadata"`
			Status            string            `json:"status"`
			CancelAtPeriodEnd bool              `json:"cancel_at_period_end"`
		}
		if err := json.Unmarshal(event.Data.Raw, &payload); err != nil {
			s.logger.Error().Err(err).Msg("Invalid customer.subscription.updated payload")
			http.Error(w, "invalid subscription data", http.StatusBadRequest)
			return
		}
		// Determine status: mark as 'cancelled' if scheduled to cancel or already canceled
		status := payload.Status
		if payload.CancelAtPeriodEnd || payload.Status == string(stripe.SubscriptionStatusCanceled) {
			status = "cancelled"
		}
		start := time.Unix(payload.CurrentPeriodStart, 0)
		end := time.Unix(payload.CurrentPeriodEnd, 0)
		planID := payload.Items.Data[0].Plan.ID
		userID := payload.Metadata["user_id"]
		if err := s.subSvc.UpsertStripeSubscription(ctx, userID, planID, start, end, status, payload.ID); err != nil {
			s.logger.Error().Err(err).Msg("Failed to update subscription on customer.subscription.updated")
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
		userID := ss.Metadata["user_id"]
		freePlanID := "free"
		start := time.Now()
		end := start.AddDate(0, 0, 31)
		if err := s.subSvc.UpsertStripeSubscription(ctx, userID, freePlanID, start, end, "active", ""); err != nil {
			s.logger.Error().Err(err).Msg("Failed to downgrade subscription on customer.subscription.deleted")
			http.Error(w, "failed to downgrade subscription", http.StatusInternalServerError)
			return
		}
	case "invoice.payment_failed":
		// Parse invoice payload for failed payment
		var payload struct {
			Customer     string            `json:"customer"`
			Subscription string            `json:"subscription"`
			PeriodStart  int64             `json:"period_start"`
			PeriodEnd    int64             `json:"period_end"`
			Metadata     map[string]string `json:"metadata"`
			Lines        struct {
				Data []struct {
					Price struct {
						ID string `json:"id"`
					} `json:"price"`
				} `json:"data"`
			} `json:"lines"`
		}
		if err := json.Unmarshal(event.Data.Raw, &payload); err != nil {
			s.logger.Error().Err(err).Msg("Invalid invoice.payment_failed payload")
			http.Error(w, "invalid invoice data", http.StatusBadRequest)
			return
		}
		userID := payload.Metadata["user_id"]
		if userID == "" {
			custID := payload.Customer
			s.logger.Warn().Str("stripe_customer_id", custID).Msg("Invoice missing user_id metadata; looking up user by customer ID")
			u, err := s.userRepo.GetUserByStripeCustomerID(ctx, custID)
			if err != nil {
				s.logger.Error().Err(err).Msg("Failed to lookup user by Stripe customer ID")
				http.Error(w, "failed to identify user", http.StatusInternalServerError)
				return
			}
			if u == nil {
				s.logger.Error().Str("stripe_customer_id", custID).Msg("No user found for customer ID")
				http.Error(w, "user not found", http.StatusBadRequest)
				return
			}
			userID = u.UserID
		}
		start := time.Unix(payload.PeriodStart, 0)
		end := time.Unix(payload.PeriodEnd, 0)
		if len(payload.Lines.Data) == 0 {
			s.logger.Warn().Msg("Invoice has no line items")
			break
		}
		priceID := payload.Lines.Data[0].Price.ID
		subID := payload.Subscription
		if err := s.subSvc.UpsertStripeSubscription(ctx, userID, priceID, start, end, "past_due", subID); err != nil {
			s.logger.Error().Err(err).Msg("Failed to mark subscription as past_due on invoice.payment_failed")
			http.Error(w, "failed to mark past_due", http.StatusInternalServerError)
			return
		}
	default:
		s.logger.Warn().Str("event_type", string(event.Type)).Msg("Unhandled Stripe webhook event")
	}
	w.WriteHeader(http.StatusOK)
}
