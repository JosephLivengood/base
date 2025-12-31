package auth

import (
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Get("/google", h.GoogleLogin)
	r.Get("/google/callback", h.GoogleCallback)
	r.Get("/logout", h.Logout)
	r.Get("/me", h.Me)
}
