export type Role = 'owner' | 'admin' | 'member'
export type InvitationStatus = 'pending' | 'accepted' | 'declined' | 'expired'

export interface Organization {
  id: string
  name: string
  slug: string
  created_at: string
  updated_at: string
}

export interface OrganizationWithRole extends Organization {
  role: Role
}

export interface Member {
  id: string
  organization_id: string
  user_id: string
  role: Role
  email: string
  name: string
  picture: string
  created_at: string
  updated_at: string
}

export interface Invitation {
  id: string
  organization_id: string
  email: string
  role: Role
  invited_by: string
  status: InvitationStatus
  expires_at: string
  created_at: string
  updated_at: string
}

export interface InvitationWithDetails extends Invitation {
  organization_name: string
  invited_by_name: string
}
