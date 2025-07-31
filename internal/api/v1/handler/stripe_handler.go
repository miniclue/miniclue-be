package handler

import (
	"net/http"

	"app/internal/service"
)

// StripeHandler handles Stripe webhook events.
type StripeHandler struct {
	stripeSvc *service.StripeService
}

// NewStripeHandler creates a new StripeHandler.
func NewStripeHandler(stripeSvc *service.StripeService) *StripeHandler {
	return &StripeHandler{stripeSvc: stripeSvc}
}

// RegisterRoutes registers the Stripe webhook endpoint.
func (h *StripeHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/stripe/webhooks", h.HandleWebhook)
}

// HandleWebhook godoc
// @Summary Process Stripe webhook events
// @Description Receives and processes Stripe webhook events for subscription management.
// @Tags stripe
// @Accept json
// @Produce json
// @Param request body string true "Stripe webhook payload"
// @Success 200 {string} string "Webhook processed successfully"
// @Failure 400 {string} string "Invalid webhook signature or payload"
// @Failure 500 {string} string "Failed to process webhook"
// @Router /stripe/webhooks [post]
func (h *StripeHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	h.stripeSvc.HandleWebhook(w, r)
}
