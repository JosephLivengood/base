package subscription

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers subscription routes under /organizations/{orgID}/subscription
func RegisterRoutes(r chi.Router, h *Handler) {
	r.Get("/", h.Get)
}
