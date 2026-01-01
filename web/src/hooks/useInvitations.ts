import { useState, useCallback, useEffect } from 'react'
import { Invitation, InvitationWithDetails, Role } from '../types/organization'

interface UseOrgInvitationsResult {
  invitations: Invitation[]
  isLoading: boolean
  error: string | null
  invite: (email: string, role: Role) => Promise<void>
  cancel: (inviteId: string) => Promise<void>
  refetch: () => Promise<void>
}

interface UseMyInvitationsResult {
  invitations: InvitationWithDetails[]
  isLoading: boolean
  error: string | null
  accept: (token: string) => Promise<void>
  decline: (token: string) => Promise<void>
  refetch: () => Promise<void>
}

export function useOrgInvitations(orgId: string | undefined): UseOrgInvitationsResult {
  const [invitations, setInvitations] = useState<Invitation[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchInvitations = useCallback(async () => {
    if (!orgId) {
      setInvitations([])
      setIsLoading(false)
      return
    }

    setIsLoading(true)
    setError(null)

    try {
      const res = await fetch(`/api/organizations/${orgId}/invitations`)
      if (!res.ok) {
        if (res.status === 403) {
          // User doesn't have permission to view invitations
          setInvitations([])
          setIsLoading(false)
          return
        }
        throw new Error('Failed to fetch invitations')
      }
      const data = await res.json()
      setInvitations(data.data || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch invitations')
      setInvitations([])
    } finally {
      setIsLoading(false)
    }
  }, [orgId])

  const invite = useCallback(async (email: string, role: Role) => {
    if (!orgId) return

    const res = await fetch(`/api/organizations/${orgId}/invitations`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, role }),
    })

    if (!res.ok) {
      const data = await res.json()
      throw new Error(data.message || 'Failed to send invitation')
    }

    const data = await res.json()
    setInvitations((prev) => [data.data, ...prev])
  }, [orgId])

  const cancel = useCallback(async (inviteId: string) => {
    if (!orgId) return

    const res = await fetch(`/api/organizations/${orgId}/invitations/${inviteId}`, {
      method: 'DELETE',
    })

    if (!res.ok) {
      const data = await res.json()
      throw new Error(data.message || 'Failed to cancel invitation')
    }

    setInvitations((prev) => prev.filter((i) => i.id !== inviteId))
  }, [orgId])

  useEffect(() => {
    fetchInvitations()
  }, [fetchInvitations])

  return {
    invitations,
    isLoading,
    error,
    invite,
    cancel,
    refetch: fetchInvitations,
  }
}

export function useMyInvitations(): UseMyInvitationsResult {
  const [invitations, setInvitations] = useState<InvitationWithDetails[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchInvitations = useCallback(async () => {
    setIsLoading(true)
    setError(null)

    try {
      const res = await fetch('/api/invitations')
      if (!res.ok) {
        throw new Error('Failed to fetch invitations')
      }
      const data = await res.json()
      setInvitations(data.data || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch invitations')
      setInvitations([])
    } finally {
      setIsLoading(false)
    }
  }, [])

  const accept = useCallback(async (token: string) => {
    const res = await fetch(`/api/invitations/${token}/accept`, {
      method: 'POST',
    })

    if (!res.ok) {
      const data = await res.json()
      throw new Error(data.message || 'Failed to accept invitation')
    }

    setInvitations((prev) => prev.filter((i) => i.id !== token))
  }, [])

  const decline = useCallback(async (token: string) => {
    const res = await fetch(`/api/invitations/${token}/decline`, {
      method: 'POST',
    })

    if (!res.ok) {
      const data = await res.json()
      throw new Error(data.message || 'Failed to decline invitation')
    }

    setInvitations((prev) => prev.filter((i) => i.id !== token))
  }, [])

  useEffect(() => {
    fetchInvitations()
  }, [fetchInvitations])

  return {
    invitations,
    isLoading,
    error,
    accept,
    decline,
    refetch: fetchInvitations,
  }
}
