package organization

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"base/api/internal/database"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrAlreadyMember   = errors.New("user is already a member")
	ErrNotMember       = errors.New("user is not a member")
	ErrLastOwner       = errors.New("cannot remove the last owner")
	ErrInviteExists    = errors.New("pending invitation already exists")
	ErrInviteExpired   = errors.New("invitation has expired")
	ErrSlugExists      = errors.New("slug already exists")
)

type Repository struct {
	postgres *database.PostgresDB
}

func NewRepository(postgres *database.PostgresDB) *Repository {
	return &Repository{postgres: postgres}
}

// Organization CRUD

func (r *Repository) Create(ctx context.Context, name, slug, createdBy string) (*Organization, error) {
	var org Organization
	query := `
		INSERT INTO organizations (name, slug, created_by)
		VALUES ($1, $2, $3)
		RETURNING id, name, slug, created_by, created_at, updated_at
	`
	err := r.postgres.GetContext(ctx, &org, query, name, slug, createdBy)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") && strings.Contains(err.Error(), "slug") {
			return nil, ErrSlugExists
		}
		return nil, err
	}
	return &org, nil
}

func (r *Repository) CreateWithOwner(ctx context.Context, name, slug, createdBy string) (*Organization, error) {
	tx, err := r.postgres.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var org Organization
	query := `
		INSERT INTO organizations (name, slug, created_by)
		VALUES ($1, $2, $3)
		RETURNING id, name, slug, created_by, created_at, updated_at
	`
	err = tx.GetContext(ctx, &org, query, name, slug, createdBy)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") && strings.Contains(err.Error(), "slug") {
			return nil, ErrSlugExists
		}
		return nil, err
	}

	memberQuery := `
		INSERT INTO organization_members (organization_id, user_id, role)
		VALUES ($1, $2, $3)
	`
	_, err = tx.ExecContext(ctx, memberQuery, org.ID, createdBy, RoleOwner)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &org, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Organization, error) {
	var org Organization
	query := `SELECT id, name, slug, created_by, created_at, updated_at FROM organizations WHERE id = $1`
	err := r.postgres.GetContext(ctx, &org, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &org, err
}

func (r *Repository) GetBySlug(ctx context.Context, slug string) (*Organization, error) {
	var org Organization
	query := `SELECT id, name, slug, created_by, created_at, updated_at FROM organizations WHERE slug = $1`
	err := r.postgres.GetContext(ctx, &org, query, slug)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &org, err
}

func (r *Repository) Update(ctx context.Context, id, name string) (*Organization, error) {
	var org Organization
	query := `
		UPDATE organizations
		SET name = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, slug, created_by, created_at, updated_at
	`
	err := r.postgres.GetContext(ctx, &org, query, id, name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &org, err
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM organizations WHERE id = $1`
	result, err := r.postgres.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) GetUserOrganizations(ctx context.Context, userID string) ([]OrganizationWithRole, error) {
	var orgs []OrganizationWithRole
	query := `
		SELECT o.id, o.name, o.slug, o.created_by, o.created_at, o.updated_at, m.role
		FROM organizations o
		JOIN organization_members m ON o.id = m.organization_id
		WHERE m.user_id = $1
		ORDER BY o.name
	`
	err := r.postgres.SelectContext(ctx, &orgs, query, userID)
	return orgs, err
}

// Member operations

func (r *Repository) AddMember(ctx context.Context, orgID, userID string, role Role) (*Member, error) {
	var member Member
	query := `
		INSERT INTO organization_members (organization_id, user_id, role)
		VALUES ($1, $2, $3)
		RETURNING id, organization_id, user_id, role, created_at, updated_at
	`
	err := r.postgres.GetContext(ctx, &member, query, orgID, userID, role)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, ErrAlreadyMember
		}
		return nil, err
	}
	return &member, nil
}

func (r *Repository) GetMember(ctx context.Context, orgID, userID string) (*Member, error) {
	var member Member
	query := `
		SELECT id, organization_id, user_id, role, created_at, updated_at
		FROM organization_members
		WHERE organization_id = $1 AND user_id = $2
	`
	err := r.postgres.GetContext(ctx, &member, query, orgID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotMember
	}
	return &member, err
}

func (r *Repository) GetMembers(ctx context.Context, orgID string) ([]MemberWithUser, error) {
	var members []MemberWithUser
	query := `
		SELECT m.id, m.organization_id, m.user_id, m.role, m.created_at, m.updated_at,
			   u.email, u.name, u.picture
		FROM organization_members m
		JOIN users u ON m.user_id = u.id
		WHERE m.organization_id = $1
		ORDER BY
			CASE m.role
				WHEN 'owner' THEN 1
				WHEN 'admin' THEN 2
				ELSE 3
			END,
			u.name
	`
	err := r.postgres.SelectContext(ctx, &members, query, orgID)
	return members, err
}

func (r *Repository) UpdateMemberRole(ctx context.Context, orgID, userID string, role Role) error {
	query := `
		UPDATE organization_members
		SET role = $3, updated_at = NOW()
		WHERE organization_id = $1 AND user_id = $2
	`
	result, err := r.postgres.ExecContext(ctx, query, orgID, userID, role)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotMember
	}
	return nil
}

func (r *Repository) RemoveMember(ctx context.Context, orgID, userID string) error {
	query := `DELETE FROM organization_members WHERE organization_id = $1 AND user_id = $2`
	result, err := r.postgres.ExecContext(ctx, query, orgID, userID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotMember
	}
	return nil
}

func (r *Repository) CountOwners(ctx context.Context, orgID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM organization_members WHERE organization_id = $1 AND role = 'owner'`
	err := r.postgres.GetContext(ctx, &count, query, orgID)
	return count, err
}

func (r *Repository) TransferOwnership(ctx context.Context, orgID, currentOwnerID, newOwnerID string) error {
	tx, err := r.postgres.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Demote current owner to admin
	demoteQuery := `
		UPDATE organization_members
		SET role = 'admin', updated_at = NOW()
		WHERE organization_id = $1 AND user_id = $2
	`
	_, err = tx.ExecContext(ctx, demoteQuery, orgID, currentOwnerID)
	if err != nil {
		return err
	}

	// Promote new owner
	promoteQuery := `
		UPDATE organization_members
		SET role = 'owner', updated_at = NOW()
		WHERE organization_id = $1 AND user_id = $2
	`
	result, err := tx.ExecContext(ctx, promoteQuery, orgID, newOwnerID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotMember
	}

	return tx.Commit()
}

// Invitation operations

func (r *Repository) CreateInvitation(ctx context.Context, orgID, email string, role Role, invitedBy string, expiresAt time.Time) (*Invitation, error) {
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	var inv Invitation
	query := `
		INSERT INTO organization_invitations (organization_id, email, role, token, invited_by, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, organization_id, email, role, token, invited_by, status, expires_at, created_at, updated_at
	`
	err = r.postgres.GetContext(ctx, &inv, query, orgID, email, role, token, invitedBy, expiresAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, ErrInviteExists
		}
		return nil, err
	}
	return &inv, nil
}

func (r *Repository) GetInvitationByToken(ctx context.Context, token string) (*InvitationWithDetails, error) {
	var inv InvitationWithDetails
	query := `
		SELECT i.id, i.organization_id, i.email, i.role, i.token, i.invited_by, i.status,
			   i.expires_at, i.created_at, i.updated_at,
			   o.name as organization_name, u.name as invited_by_name
		FROM organization_invitations i
		JOIN organizations o ON i.organization_id = o.id
		JOIN users u ON i.invited_by = u.id
		WHERE i.token = $1
	`
	err := r.postgres.GetContext(ctx, &inv, query, token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &inv, err
}

func (r *Repository) GetInvitationByID(ctx context.Context, id string) (*Invitation, error) {
	var inv Invitation
	query := `
		SELECT id, organization_id, email, role, token, invited_by, status, expires_at, created_at, updated_at
		FROM organization_invitations
		WHERE id = $1
	`
	err := r.postgres.GetContext(ctx, &inv, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &inv, err
}

func (r *Repository) GetPendingInvitationsForOrg(ctx context.Context, orgID string) ([]Invitation, error) {
	var invitations []Invitation
	query := `
		SELECT id, organization_id, email, role, token, invited_by, status, expires_at, created_at, updated_at
		FROM organization_invitations
		WHERE organization_id = $1 AND status = 'pending'
		ORDER BY created_at DESC
	`
	err := r.postgres.SelectContext(ctx, &invitations, query, orgID)
	return invitations, err
}

func (r *Repository) GetPendingInvitationsForEmail(ctx context.Context, email string) ([]InvitationWithDetails, error) {
	var invitations []InvitationWithDetails
	query := `
		SELECT i.id, i.organization_id, i.email, i.role, i.token, i.invited_by, i.status,
			   i.expires_at, i.created_at, i.updated_at,
			   o.name as organization_name, u.name as invited_by_name
		FROM organization_invitations i
		JOIN organizations o ON i.organization_id = o.id
		JOIN users u ON i.invited_by = u.id
		WHERE i.email = $1 AND i.status = 'pending' AND i.expires_at > NOW()
		ORDER BY i.created_at DESC
	`
	err := r.postgres.SelectContext(ctx, &invitations, query, email)
	return invitations, err
}

func (r *Repository) UpdateInvitationStatus(ctx context.Context, id string, status InvitationStatus) error {
	query := `
		UPDATE organization_invitations
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`
	result, err := r.postgres.ExecContext(ctx, query, id, status)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteInvitation(ctx context.Context, id string) error {
	query := `DELETE FROM organization_invitations WHERE id = $1`
	result, err := r.postgres.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) IsMemberByEmail(ctx context.Context, orgID, email string) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM organization_members m
		JOIN users u ON m.user_id = u.id
		WHERE m.organization_id = $1 AND u.email = $2
	`
	err := r.postgres.GetContext(ctx, &count, query, orgID, email)
	return count > 0, err
}

// Helper functions

func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func GenerateSlug(name, userID string) string {
	slug := strings.ToLower(name)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if len(slug) > 30 {
		slug = slug[:30]
	}
	// Append part of userID for uniqueness
	if len(userID) >= 8 {
		slug = fmt.Sprintf("%s-%s", slug, userID[:8])
	}
	return slug
}
