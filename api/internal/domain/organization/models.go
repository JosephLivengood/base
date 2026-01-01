package organization

import "time"

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleMember:
		return true
	}
	return false
}

func (r Role) CanManageMembers() bool {
	return r == RoleOwner || r == RoleAdmin
}

func (r Role) CanDeleteOrg() bool {
	return r == RoleOwner
}

func (r Role) CanTransferOwnership() bool {
	return r == RoleOwner
}

type InvitationStatus string

const (
	StatusPending  InvitationStatus = "pending"
	StatusAccepted InvitationStatus = "accepted"
	StatusDeclined InvitationStatus = "declined"
	StatusExpired  InvitationStatus = "expired"
)

type Organization struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Slug      string    `json:"slug" db:"slug"`
	CreatedBy string    `json:"-" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type OrganizationWithRole struct {
	Organization
	Role Role `json:"role" db:"role"`
}

type Member struct {
	ID             string    `json:"id" db:"id"`
	OrganizationID string    `json:"organization_id" db:"organization_id"`
	UserID         string    `json:"user_id" db:"user_id"`
	Role           Role      `json:"role" db:"role"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type MemberWithUser struct {
	ID             string    `json:"id" db:"id"`
	OrganizationID string    `json:"organization_id" db:"organization_id"`
	UserID         string    `json:"user_id" db:"user_id"`
	Role           Role      `json:"role" db:"role"`
	Email          string    `json:"email" db:"email"`
	Name           string    `json:"name" db:"name"`
	Picture        string    `json:"picture" db:"picture"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type Invitation struct {
	ID             string           `json:"id" db:"id"`
	OrganizationID string           `json:"organization_id" db:"organization_id"`
	Email          string           `json:"email" db:"email"`
	Role           Role             `json:"role" db:"role"`
	Token          string           `json:"-" db:"token"`
	InvitedBy      string           `json:"invited_by" db:"invited_by"`
	Status         InvitationStatus `json:"status" db:"status"`
	ExpiresAt      time.Time        `json:"expires_at" db:"expires_at"`
	CreatedAt      time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at" db:"updated_at"`
}

type InvitationWithDetails struct {
	Invitation
	OrganizationName string `json:"organization_name" db:"organization_name"`
	InvitedByName    string `json:"invited_by_name" db:"invited_by_name"`
}

// Request types

type CreateOrgRequest struct {
	Name string `json:"name"`
}

type UpdateOrgRequest struct {
	Name string `json:"name"`
}

type InviteMemberRequest struct {
	Email string `json:"email"`
	Role  Role   `json:"role"`
}

type UpdateMemberRoleRequest struct {
	Role Role `json:"role"`
}

type TransferOwnershipRequest struct {
	NewOwnerID string `json:"new_owner_id"`
}

type SetActiveOrgRequest struct {
	OrganizationID string `json:"organization_id"`
}
