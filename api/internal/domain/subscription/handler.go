package subscription

import (
	"errors"
	"net/http"

	"base/api/internal/domain/organization"
	"base/api/internal/middleware"
	"base/api/pkg/response"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	repo    *Repository
	orgRepo *organization.Repository
}

func NewHandler(repo *Repository, orgRepo *organization.Repository) *Handler {
	return &Handler{
		repo:    repo,
		orgRepo: orgRepo,
	}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	orgID := chi.URLParam(r, "orgID")

	// Verify user is a member of the org
	_, err := h.orgRepo.GetMember(r.Context(), orgID, usr.ID)
	if err != nil {
		if errors.Is(err, organization.ErrNotMember) {
			response.Forbidden(w, "not a member of this organization")
			return
		}
		response.InternalError(w, "failed to check membership")
		return
	}

	sub, err := h.repo.GetActiveByOrganization(r.Context(), orgID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// No active subscription - return null
			response.OK(w, nil)
			return
		}
		response.InternalError(w, "failed to get subscription")
		return
	}

	response.OK(w, sub)
}
