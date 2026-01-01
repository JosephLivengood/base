import { useState, useCallback, useEffect } from 'react'
import { Member, Role } from '../types/organization'

interface UseMembersResult {
  members: Member[]
  isLoading: boolean
  error: string | null
  updateRole: (userId: string, role: Role) => Promise<void>
  removeMember: (userId: string) => Promise<void>
  refetch: () => Promise<void>
}

export function useMembers(orgId: string | undefined): UseMembersResult {
  const [members, setMembers] = useState<Member[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchMembers = useCallback(async () => {
    if (!orgId) {
      setMembers([])
      setIsLoading(false)
      return
    }

    setIsLoading(true)
    setError(null)

    try {
      const res = await fetch(`/api/organizations/${orgId}/members`)
      if (!res.ok) {
        throw new Error('Failed to fetch members')
      }
      const data = await res.json()
      setMembers(data.data || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch members')
      setMembers([])
    } finally {
      setIsLoading(false)
    }
  }, [orgId])

  const updateRole = useCallback(async (userId: string, role: Role) => {
    if (!orgId) return

    const res = await fetch(`/api/organizations/${orgId}/members/${userId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ role }),
    })

    if (!res.ok) {
      const data = await res.json()
      throw new Error(data.message || 'Failed to update role')
    }

    // Update local state
    setMembers((prev) =>
      prev.map((m) => (m.user_id === userId ? { ...m, role } : m))
    )
  }, [orgId])

  const removeMember = useCallback(async (userId: string) => {
    if (!orgId) return

    const res = await fetch(`/api/organizations/${orgId}/members/${userId}`, {
      method: 'DELETE',
    })

    if (!res.ok) {
      const data = await res.json()
      throw new Error(data.message || 'Failed to remove member')
    }

    // Update local state
    setMembers((prev) => prev.filter((m) => m.user_id !== userId))
  }, [orgId])

  useEffect(() => {
    fetchMembers()
  }, [fetchMembers])

  return {
    members,
    isLoading,
    error,
    updateRole,
    removeMember,
    refetch: fetchMembers,
  }
}
