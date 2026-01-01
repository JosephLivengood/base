import { useState } from 'react'
import { useAuth } from '../../hooks/useAuth'
import { useMembers } from '../../hooks/useMembers'
import { Role } from '../../types/organization'

interface MemberListProps {
  orgId: string
  currentUserRole: Role
}

export function MemberList({ orgId, currentUserRole }: MemberListProps) {
  const { user } = useAuth()
  const { members, isLoading, error, updateRole, removeMember } = useMembers(orgId)
  const [actionError, setActionError] = useState<string | null>(null)

  const canManageMembers = currentUserRole === 'owner' || currentUserRole === 'admin'

  const handleRoleChange = async (userId: string, newRole: Role) => {
    setActionError(null)
    try {
      await updateRole(userId, newRole)
    } catch (err) {
      setActionError(err instanceof Error ? err.message : 'Failed to update role')
    }
  }

  const handleRemove = async (userId: string, memberName: string) => {
    if (!confirm(`Are you sure you want to remove ${memberName} from this organization?`)) {
      return
    }

    setActionError(null)
    try {
      await removeMember(userId)
    } catch (err) {
      setActionError(err instanceof Error ? err.message : 'Failed to remove member')
    }
  }

  if (isLoading) {
    return <div className="text-gray-400">Loading members...</div>
  }

  if (error) {
    return <div className="text-red-400">{error}</div>
  }

  return (
    <div>
      {actionError && (
        <div className="mb-4 p-3 bg-red-900/50 border border-red-700 rounded-md text-red-300 text-sm">
          {actionError}
        </div>
      )}
      <div className="bg-gray-800 rounded-lg overflow-hidden">
        <ul className="divide-y divide-gray-700">
          {members.map((member) => {
            const isCurrentUser = member.user_id === user?.id
            const canModify = canManageMembers && !isCurrentUser &&
              (currentUserRole === 'owner' || member.role !== 'owner')

            return (
              <li key={member.id} className="px-4 py-3 flex items-center justify-between">
                <div className="flex items-center gap-3">
                  {member.picture ? (
                    <img
                      src={member.picture}
                      alt={member.name}
                      className="w-10 h-10 rounded-full"
                    />
                  ) : (
                    <div className="w-10 h-10 rounded-full bg-gray-700 flex items-center justify-center text-gray-400">
                      {member.name.charAt(0).toUpperCase()}
                    </div>
                  )}
                  <div>
                    <p className="text-white font-medium">
                      {member.name}
                      {isCurrentUser && (
                        <span className="ml-2 text-xs text-gray-500">(you)</span>
                      )}
                    </p>
                    <p className="text-sm text-gray-400">{member.email}</p>
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  {canModify ? (
                    <select
                      value={member.role}
                      onChange={(e) => handleRoleChange(member.user_id, e.target.value as Role)}
                      className="px-2 py-1 text-sm bg-gray-700 border border-gray-600 rounded text-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      {currentUserRole === 'owner' && (
                        <option value="owner">Owner</option>
                      )}
                      <option value="admin">Admin</option>
                      <option value="member">Member</option>
                    </select>
                  ) : (
                    <span className="px-2 py-1 text-sm bg-gray-700 text-gray-300 rounded capitalize">
                      {member.role}
                    </span>
                  )}
                  {canModify && (
                    <button
                      onClick={() => handleRemove(member.user_id, member.name)}
                      className="p-1 text-gray-400 hover:text-red-400 transition-colors"
                      title="Remove member"
                    >
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  )}
                </div>
              </li>
            )
          })}
        </ul>
      </div>
    </div>
  )
}
