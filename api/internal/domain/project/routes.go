package project

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers project routes under /organizations/{orgID}/projects
func RegisterRoutes(r chi.Router, h *Handler) {
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{projectID}", h.Get)
	r.Put("/{projectID}", h.Update)
	r.Delete("/{projectID}", h.Delete)
}
