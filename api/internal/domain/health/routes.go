package health

import (
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Get("/", h.Check)
	r.Get("/ready", h.Ready)
}
