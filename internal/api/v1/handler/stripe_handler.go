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

// HandleWebhook processes incoming Stripe webhooks.
func (h *StripeHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	h.stripeSvc.HandleWebhook(w, r)
}
