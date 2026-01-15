package subscription

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers subscription routes under /organizations/{orgID}/subscription
func RegisterRoutes(r chi.Router, h *Handler) {
	r.Get("/", h.Get)
	r.Post("/checkout", h.CreateCheckout)
	r.Post("/portal", h.CreatePortalSession)
	r.Post("/confirm", h.ConfirmCheckout)
}

// RegisterWebhookRoute registers the Stripe webhook route (no auth required)
func RegisterWebhookRoute(r chi.Router, h *Handler) {
	r.Post("/webhook", h.HandleWebhook)
}
