package organization

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"base/api/internal/domain/user"
	"base/api/internal/middleware"
	"base/api/internal/session"
	"base/api/pkg/response"

	"github.com/go-chi/chi/v5"
)

const invitationExpiryDays = 7

type Handler struct {
	repo         *Repository
	userRepo     *user.Repository
	sessionStore *session.Store
}

func NewHandler(repo *Repository, userRepo *user.Repository, sessionStore *session.Store) *Handler {
	return &Handler{
		repo:         repo,
		userRepo:     userRepo,
		sessionStore: sessionStore,
	}
}

// Organization handlers

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	if usr == nil {
		response.Unauthorized(w, "authentication required")
		return
	}

	var req CreateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.Name == "" {
		response.BadRequest(w, "name is required")
		return
	}

	slug := GenerateSlug(req.Name, usr.ID)
	org, err := h.repo.CreateWithOwner(r.Context(), req.Name, slug, usr.ID)
	if err != nil {
		if errors.Is(err, ErrSlugExists) {
			response.BadRequest(w, "organization with similar name already exists")
			return
		}
		response.InternalError(w, "failed to create organization")
		return
	}

	response.Created(w, OrganizationWithRole{
		Organization: *org,
		Role:         RoleOwner,
	})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	if usr == nil {
		response.Unauthorized(w, "authentication required")
		return
	}

	orgs, err := h.repo.GetUserOrganizations(r.Context(), usr.ID)
	if err != nil {
		response.InternalError(w, "failed to list organizations")
		return
	}

	response.OK(w, orgs)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	orgID := chi.URLParam(r, "orgID")

	member, err := h.repo.GetMember(r.Context(), orgID, usr.ID)
	if err != nil {
		if errors.Is(err, ErrNotMember) {
			response.Forbidden(w, "not a member of this organization")
			return
		}
		response.InternalError(w, "failed to check membership")
		return
	}

	org, err := h.repo.GetByID(r.Context(), orgID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, "organization not found")
			return
		}
		response.InternalError(w, "failed to get organization")
		return
	}

	response.OK(w, OrganizationWithRole{
		Organization: *org,
		Role:         member.Role,
	})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	orgID := chi.URLParam(r, "orgID")

	member, err := h.repo.GetMember(r.Context(), orgID, usr.ID)
	if err != nil {
		if errors.Is(err, ErrNotMember) {
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

	var req UpdateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.Name == "" {
		response.BadRequest(w, "name is required")
		return
	}

	org, err := h.repo.Update(r.Context(), orgID, req.Name)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, "organization not found")
			return
		}
		response.InternalError(w, "failed to update organization")
		return
	}

	response.OK(w, OrganizationWithRole{
		Organization: *org,
		Role:         member.Role,
	})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	orgID := chi.URLParam(r, "orgID")

	member, err := h.repo.GetMember(r.Context(), orgID, usr.ID)
	if err != nil {
		if errors.Is(err, ErrNotMember) {
			response.Forbidden(w, "not a member of this organization")
			return
		}
		response.InternalError(w, "failed to check membership")
		return
	}

	if !member.Role.CanDeleteOrg() {
		response.Forbidden(w, "only owners can delete organizations")
		return
	}

	if err := h.repo.Delete(r.Context(), orgID); err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, "organization not found")
			return
		}
		response.InternalError(w, "failed to delete organization")
		return
	}

	response.NoContent(w)
}

// Member handlers

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	orgID := chi.URLParam(r, "orgID")

	_, err := h.repo.GetMember(r.Context(), orgID, usr.ID)
	if err != nil {
		if errors.Is(err, ErrNotMember) {
			response.Forbidden(w, "not a member of this organization")
			return
		}
		response.InternalError(w, "failed to check membership")
		return
	}

	members, err := h.repo.GetMembers(r.Context(), orgID)
	if err != nil {
		response.InternalError(w, "failed to list members")
		return
	}

	response.OK(w, members)
}

func (h *Handler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	orgID := chi.URLParam(r, "orgID")
	targetUserID := chi.URLParam(r, "userID")

	member, err := h.repo.GetMember(r.Context(), orgID, usr.ID)
	if err != nil {
		if errors.Is(err, ErrNotMember) {
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

	var req UpdateMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if !req.Role.IsValid() {
		response.BadRequest(w, "invalid role")
		return
	}

	targetMember, err := h.repo.GetMember(r.Context(), orgID, targetUserID)
	if err != nil {
		if errors.Is(err, ErrNotMember) {
			response.NotFound(w, "member not found")
			return
		}
		response.InternalError(w, "failed to get member")
		return
	}

	// Only owners can promote to owner or demote owners
	if req.Role == RoleOwner || targetMember.Role == RoleOwner {
		if member.Role != RoleOwner {
			response.Forbidden(w, "only owners can change owner roles")
			return
		}
	}

	// Check if demoting the last owner
	if targetMember.Role == RoleOwner && req.Role != RoleOwner {
		count, err := h.repo.CountOwners(r.Context(), orgID)
		if err != nil {
			response.InternalError(w, "failed to check owners")
			return
		}
		if count <= 1 {
			response.BadRequest(w, "cannot demote the last owner")
			return
		}
	}

	if err := h.repo.UpdateMemberRole(r.Context(), orgID, targetUserID, req.Role); err != nil {
		response.InternalError(w, "failed to update role")
		return
	}

	response.NoContent(w)
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	orgID := chi.URLParam(r, "orgID")
	targetUserID := chi.URLParam(r, "userID")

	member, err := h.repo.GetMember(r.Context(), orgID, usr.ID)
	if err != nil {
		if errors.Is(err, ErrNotMember) {
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

	targetMember, err := h.repo.GetMember(r.Context(), orgID, targetUserID)
	if err != nil {
		if errors.Is(err, ErrNotMember) {
			response.NotFound(w, "member not found")
			return
		}
		response.InternalError(w, "failed to get member")
		return
	}

	// Admins cannot remove owners
	if targetMember.Role == RoleOwner && member.Role != RoleOwner {
		response.Forbidden(w, "admins cannot remove owners")
		return
	}

	// Prevent removing the last owner
	if targetMember.Role == RoleOwner {
		count, err := h.repo.CountOwners(r.Context(), orgID)
		if err != nil {
			response.InternalError(w, "failed to check owners")
			return
		}
		if count <= 1 {
			response.BadRequest(w, "cannot remove the last owner")
			return
		}
	}

	if err := h.repo.RemoveMember(r.Context(), orgID, targetUserID); err != nil {
		response.InternalError(w, "failed to remove member")
		return
	}

	response.NoContent(w)
}

func (h *Handler) Leave(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	orgID := chi.URLParam(r, "orgID")

	member, err := h.repo.GetMember(r.Context(), orgID, usr.ID)
	if err != nil {
		if errors.Is(err, ErrNotMember) {
			response.Forbidden(w, "not a member of this organization")
			return
		}
		response.InternalError(w, "failed to check membership")
		return
	}

	// Check if sole owner
	if member.Role == RoleOwner {
		count, err := h.repo.CountOwners(r.Context(), orgID)
		if err != nil {
			response.InternalError(w, "failed to check owners")
			return
		}
		if count <= 1 {
			response.BadRequest(w, "cannot leave as the last owner - transfer ownership or delete the organization")
			return
		}
	}

	if err := h.repo.RemoveMember(r.Context(), orgID, usr.ID); err != nil {
		response.InternalError(w, "failed to leave organization")
		return
	}

	response.NoContent(w)
}

func (h *Handler) TransferOwnership(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	orgID := chi.URLParam(r, "orgID")

	member, err := h.repo.GetMember(r.Context(), orgID, usr.ID)
	if err != nil {
		if errors.Is(err, ErrNotMember) {
			response.Forbidden(w, "not a member of this organization")
			return
		}
		response.InternalError(w, "failed to check membership")
		return
	}

	if !member.Role.CanTransferOwnership() {
		response.Forbidden(w, "only owners can transfer ownership")
		return
	}

	var req TransferOwnershipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.NewOwnerID == "" {
		response.BadRequest(w, "new_owner_id is required")
		return
	}

	if req.NewOwnerID == usr.ID {
		response.BadRequest(w, "cannot transfer ownership to yourself")
		return
	}

	// Verify new owner is a member
	_, err = h.repo.GetMember(r.Context(), orgID, req.NewOwnerID)
	if err != nil {
		if errors.Is(err, ErrNotMember) {
			response.BadRequest(w, "new owner must be a member of the organization")
			return
		}
		response.InternalError(w, "failed to verify new owner")
		return
	}

	if err := h.repo.TransferOwnership(r.Context(), orgID, usr.ID, req.NewOwnerID); err != nil {
		response.InternalError(w, "failed to transfer ownership")
		return
	}

	response.NoContent(w)
}

// Invitation handlers

func (h *Handler) Invite(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	orgID := chi.URLParam(r, "orgID")

	member, err := h.repo.GetMember(r.Context(), orgID, usr.ID)
	if err != nil {
		if errors.Is(err, ErrNotMember) {
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

	var req InviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.Email == "" {
		response.BadRequest(w, "email is required")
		return
	}

	if !req.Role.IsValid() || req.Role == RoleOwner {
		response.BadRequest(w, "invalid role - must be admin or member")
		return
	}

	// Check if already a member
	isMember, err := h.repo.IsMemberByEmail(r.Context(), orgID, req.Email)
	if err != nil {
		response.InternalError(w, "failed to check membership")
		return
	}
	if isMember {
		response.BadRequest(w, "user is already a member of this organization")
		return
	}

	expiresAt := time.Now().Add(invitationExpiryDays * 24 * time.Hour)
	inv, err := h.repo.CreateInvitation(r.Context(), orgID, req.Email, req.Role, usr.ID, expiresAt)
	if err != nil {
		if errors.Is(err, ErrInviteExists) {
			response.BadRequest(w, "pending invitation already exists for this email")
			return
		}
		response.InternalError(w, "failed to create invitation")
		return
	}

	response.Created(w, inv)
}

func (h *Handler) ListInvitations(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	orgID := chi.URLParam(r, "orgID")

	member, err := h.repo.GetMember(r.Context(), orgID, usr.ID)
	if err != nil {
		if errors.Is(err, ErrNotMember) {
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

	invitations, err := h.repo.GetPendingInvitationsForOrg(r.Context(), orgID)
	if err != nil {
		response.InternalError(w, "failed to list invitations")
		return
	}

	response.OK(w, invitations)
}

func (h *Handler) CancelInvitation(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	orgID := chi.URLParam(r, "orgID")
	inviteID := chi.URLParam(r, "inviteID")

	member, err := h.repo.GetMember(r.Context(), orgID, usr.ID)
	if err != nil {
		if errors.Is(err, ErrNotMember) {
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

	// Verify invitation belongs to this org
	inv, err := h.repo.GetInvitationByID(r.Context(), inviteID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, "invitation not found")
			return
		}
		response.InternalError(w, "failed to get invitation")
		return
	}

	if inv.OrganizationID != orgID {
		response.NotFound(w, "invitation not found")
		return
	}

	if err := h.repo.DeleteInvitation(r.Context(), inviteID); err != nil {
		response.InternalError(w, "failed to cancel invitation")
		return
	}

	response.NoContent(w)
}

// User invitation handlers (for invitations sent TO the current user)

func (h *Handler) MyInvitations(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	if usr == nil {
		response.Unauthorized(w, "authentication required")
		return
	}

	invitations, err := h.repo.GetPendingInvitationsForEmail(r.Context(), usr.Email)
	if err != nil {
		response.InternalError(w, "failed to list invitations")
		return
	}

	response.OK(w, invitations)
}

func (h *Handler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	if usr == nil {
		response.Unauthorized(w, "authentication required")
		return
	}

	token := chi.URLParam(r, "token")

	inv, err := h.repo.GetInvitationByToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, "invitation not found")
			return
		}
		response.InternalError(w, "failed to get invitation")
		return
	}

	// Verify invitation is for this user
	if inv.Email != usr.Email {
		response.Forbidden(w, "invitation is not for this user")
		return
	}

	// Check if expired
	if time.Now().After(inv.ExpiresAt) {
		h.repo.UpdateInvitationStatus(r.Context(), inv.ID, StatusExpired)
		response.BadRequest(w, "invitation has expired")
		return
	}

	if inv.Status != StatusPending {
		response.BadRequest(w, "invitation is no longer pending")
		return
	}

	// Add user as member
	_, err = h.repo.AddMember(r.Context(), inv.OrganizationID, usr.ID, inv.Role)
	if err != nil {
		if errors.Is(err, ErrAlreadyMember) {
			// Already a member, just mark invitation as accepted
			h.repo.UpdateInvitationStatus(r.Context(), inv.ID, StatusAccepted)
			response.BadRequest(w, "already a member of this organization")
			return
		}
		response.InternalError(w, "failed to add member")
		return
	}

	// Mark invitation as accepted
	if err := h.repo.UpdateInvitationStatus(r.Context(), inv.ID, StatusAccepted); err != nil {
		response.InternalError(w, "failed to update invitation status")
		return
	}

	// Return the organization
	org, err := h.repo.GetByID(r.Context(), inv.OrganizationID)
	if err != nil {
		response.InternalError(w, "failed to get organization")
		return
	}

	response.OK(w, OrganizationWithRole{
		Organization: *org,
		Role:         inv.Role,
	})
}

func (h *Handler) DeclineInvitation(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	if usr == nil {
		response.Unauthorized(w, "authentication required")
		return
	}

	token := chi.URLParam(r, "token")

	inv, err := h.repo.GetInvitationByToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, "invitation not found")
			return
		}
		response.InternalError(w, "failed to get invitation")
		return
	}

	// Verify invitation is for this user
	if inv.Email != usr.Email {
		response.Forbidden(w, "invitation is not for this user")
		return
	}

	if inv.Status != StatusPending {
		response.BadRequest(w, "invitation is no longer pending")
		return
	}

	if err := h.repo.UpdateInvitationStatus(r.Context(), inv.ID, StatusDeclined); err != nil {
		response.InternalError(w, "failed to decline invitation")
		return
	}

	response.NoContent(w)
}

// SetActiveOrg sets the active organization for the current session
func (h *Handler) SetActiveOrg(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUserFromContext(r.Context())
	sess := middleware.GetSessionFromContext(r.Context())
	if usr == nil || sess == nil {
		response.Unauthorized(w, "authentication required")
		return
	}

	var req SetActiveOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.OrganizationID == "" {
		response.BadRequest(w, "organization_id is required")
		return
	}

	// Verify user is a member of the organization
	_, err := h.repo.GetMember(r.Context(), req.OrganizationID, usr.ID)
	if err != nil {
		if errors.Is(err, ErrNotMember) {
			response.Forbidden(w, "not a member of this organization")
			return
		}
		response.InternalError(w, "failed to check membership")
		return
	}

	// Set active org in session
	if err := h.sessionStore.SetActiveOrg(r.Context(), sess.ID, req.OrganizationID); err != nil {
		response.InternalError(w, "failed to set active organization")
		return
	}

	response.NoContent(w)
}
