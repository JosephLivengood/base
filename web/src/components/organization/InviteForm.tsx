import { useState } from 'react'
import { useOrgInvitations } from '../../hooks/useInvitations'
import { Role } from '../../types/organization'

interface InviteFormProps {
  orgId: string
}

export function InviteForm({ orgId }: InviteFormProps) {
  const { invitations, isLoading, invite, cancel } = useOrgInvitations(orgId)
  const [email, setEmail] = useState('')
  const [role, setRole] = useState<Role>('member')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!email.trim()) return

    setIsSubmitting(true)
    setError(null)
    setSuccess(null)

    try {
      await invite(email, role)
      setSuccess(`Invitation sent to ${email}`)
      setEmail('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to send invitation')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleCancel = async (inviteId: string) => {
    try {
      await cancel(inviteId)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to cancel invitation')
    }
  }

  return (
    <div className="space-y-6">
      <form onSubmit={handleSubmit} className="bg-gray-800 rounded-lg p-4">
        <h3 className="text-lg font-medium text-white mb-4">Invite Member</h3>
        <div className="flex gap-3">
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="Email address"
            className="flex-1 px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
          <select
            value={role}
            onChange={(e) => setRole(e.target.value as Role)}
            className="px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="member">Member</option>
            <option value="admin">Admin</option>
          </select>
          <button
            type="submit"
            disabled={isSubmitting || !email.trim()}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white text-sm font-medium rounded-md transition-colors"
          >
            {isSubmitting ? 'Sending...' : 'Invite'}
          </button>
        </div>
        {error && <p className="mt-2 text-sm text-red-400">{error}</p>}
        {success && <p className="mt-2 text-sm text-green-400">{success}</p>}
      </form>

      {!isLoading && invitations.length > 0 && (
        <div>
          <h3 className="text-lg font-medium text-white mb-3">Pending Invitations</h3>
          <div className="bg-gray-800 rounded-lg overflow-hidden">
            <ul className="divide-y divide-gray-700">
              {invitations.map((inv) => (
                <li key={inv.id} className="px-4 py-3 flex items-center justify-between">
                  <div>
                    <p className="text-white">{inv.email}</p>
                    <p className="text-sm text-gray-400">
                      Role: <span className="capitalize">{inv.role}</span>
                      {' | '}
                      Expires: {new Date(inv.expires_at).toLocaleDateString()}
                    </p>
                  </div>
                  <button
                    onClick={() => handleCancel(inv.id)}
                    className="text-sm text-red-400 hover:text-red-300 transition-colors"
                  >
                    Cancel
                  </button>
                </li>
              ))}
            </ul>
          </div>
        </div>
      )}
    </div>
  )
}
