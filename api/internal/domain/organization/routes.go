package organization

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers organization routes
// All routes require authentication (applied at router level)
func RegisterRoutes(r chi.Router, h *Handler) {
	// Organization CRUD
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Put("/active", h.SetActiveOrg)

	r.Route("/{orgID}", func(r chi.Router) {
		r.Get("/", h.Get)
		r.Put("/", h.Update)
		r.Delete("/", h.Delete)
		r.Post("/leave", h.Leave)
		r.Post("/transfer", h.TransferOwnership)

		// Members
		r.Get("/members", h.ListMembers)
		r.Put("/members/{userID}", h.UpdateMemberRole)
		r.Delete("/members/{userID}", h.RemoveMember)

		// Invitations (org-scoped)
		r.Post("/invitations", h.Invite)
		r.Get("/invitations", h.ListInvitations)
		r.Delete("/invitations/{inviteID}", h.CancelInvitation)
	})
}

// RegisterInvitationRoutes registers user invitation routes
// These are for invitations sent TO the current user
func RegisterInvitationRoutes(r chi.Router, h *Handler) {
	r.Get("/", h.MyInvitations)
	r.Post("/{token}/accept", h.AcceptInvitation)
	r.Post("/{token}/decline", h.DeclineInvitation)
}
