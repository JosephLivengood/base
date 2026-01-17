package project

import (
	"errors"
	"net/http"

	"base/api/internal/domain/organization"
	"base/api/internal/middleware"
	"base/api/pkg/response"
	"base/api/pkg/validate"

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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req CreateProjectRequest
	if err := validate.Request(r, &req); err != nil {
		response.BadRequest(w, validate.FirstError(err))
		return
	}

	project, err := h.repo.Create(r.Context(), orgID, req.Name)
	if err != nil {
		response.InternalError(w, "failed to create project")
		return
	}

	response.Created(w, project)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
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

	projects, err := h.repo.GetByOrganization(r.Context(), orgID)
	if err != nil {
		response.InternalError(w, "failed to list projects")
		return
	}

	response.OK(w, projects)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	orgID := chi.URLParam(r, "orgID")
	projectID := chi.URLParam(r, "projectID")

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

	project, err := h.repo.GetByID(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, "project not found")
			return
		}
		response.InternalError(w, "failed to get project")
		return
	}

	// Verify project belongs to this org
	if project.OrganizationID != orgID {
		response.NotFound(w, "project not found")
		return
	}

	response.OK(w, project)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	orgID := chi.URLParam(r, "orgID")
	projectID := chi.URLParam(r, "projectID")

	// Verify user is admin+ in the org
	member, err := h.orgRepo.GetMember(r.Context(), orgID, usr.ID)
	if err != nil {
		if errors.Is(err, organization.ErrNotMember) {
			response.Forbidden(w, "not a member of this organization")
			return
		}
		response.InternalError(w, "failed to check membership")
		return
	}

	if !member.Role.CanManageMembers() {
		response.Forbidden(w, "insufficient permissions")
		return
	}

	// Verify project belongs to this org
	existing, err := h.repo.GetByID(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, "project not found")
			return
		}
		response.InternalError(w, "failed to get project")
		return
	}

	if existing.OrganizationID != orgID {
		response.NotFound(w, "project not found")
		return
	}

	var req UpdateProjectRequest
	if err := validate.Request(r, &req); err != nil {
		response.BadRequest(w, validate.FirstError(err))
		return
	}

	project, err := h.repo.Update(r.Context(), projectID, req.Name)
	if err != nil {
		response.InternalError(w, "failed to update project")
		return
	}

	response.OK(w, project)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	orgID := chi.URLParam(r, "orgID")
	projectID := chi.URLParam(r, "projectID")

	// Verify user is admin+ in the org
	member, err := h.orgRepo.GetMember(r.Context(), orgID, usr.ID)
	if err != nil {
		if errors.Is(err, organization.ErrNotMember) {
			response.Forbidden(w, "not a member of this organization")
			return
		}
		response.InternalError(w, "failed to check membership")
		return
	}

	if !member.Role.CanManageMembers() {
		response.Forbidden(w, "insufficient permissions")
		return
	}

	// Verify project belongs to this org
	existing, err := h.repo.GetByID(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, "project not found")
			return
		}
		response.InternalError(w, "failed to get project")
		return
	}

	if existing.OrganizationID != orgID {
		response.NotFound(w, "project not found")
		return
	}

	if err := h.repo.Delete(r.Context(), projectID); err != nil {
		response.InternalError(w, "failed to delete project")
		return
	}

	response.NoContent(w)
}
