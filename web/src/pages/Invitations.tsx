import { useState } from 'react'
import { useMyInvitations } from '../hooks/useInvitations'
import { useOrganization } from '../hooks/useOrganization'

export function Invitations() {
  const { invitations, isLoading, accept, decline, refetch } = useMyInvitations()
  const { refetch: refetchOrgs } = useOrganization()
  const [processing, setProcessing] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const handleAccept = async (token: string) => {
    setProcessing(token)
    setError(null)
    try {
      await accept(token)
      await refetchOrgs()
      await refetch()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to accept invitation')
    } finally {
      setProcessing(null)
    }
  }

  const handleDecline = async (token: string) => {
    setProcessing(token)
    setError(null)
    try {
      await decline(token)
      await refetch()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to decline invitation')
    } finally {
      setProcessing(null)
    }
  }

  if (isLoading) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-8">
        <div className="text-gray-400">Loading invitations...</div>
      </div>
    )
  }

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold text-white mb-6">Pending Invitations</h1>

      {error && (
        <div className="mb-6 p-3 bg-red-900/50 border border-red-700 rounded-md text-red-300 text-sm">
          {error}
        </div>
      )}

      {invitations.length === 0 ? (
        <div className="bg-gray-800 rounded-lg px-6 py-12 text-center text-gray-400">
          No pending invitations
        </div>
      ) : (
        <div className="bg-gray-800 rounded-lg overflow-hidden">
          <ul className="divide-y divide-gray-700">
            {invitations.map((inv) => (
              <li key={inv.id} className="px-6 py-4">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-white font-medium">{inv.organization_name}</h3>
                    <p className="text-sm text-gray-400">
                      Invited by {inv.invited_by_name} as{' '}
                      <span className="capitalize">{inv.role}</span>
                    </p>
                    <p className="text-xs text-gray-500 mt-1">
                      Expires {new Date(inv.expires_at).toLocaleDateString()}
                    </p>
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={() => handleDecline(inv.id)}
                      disabled={processing === inv.id}
                      className="px-3 py-1.5 text-sm text-gray-300 hover:text-white border border-gray-600 hover:border-gray-500 rounded-md transition-colors disabled:opacity-50"
                    >
                      Decline
                    </button>
                    <button
                      onClick={() => handleAccept(inv.id)}
                      disabled={processing === inv.id}
                      className="px-3 py-1.5 text-sm bg-blue-600 hover:bg-blue-700 text-white rounded-md transition-colors disabled:opacity-50"
                    >
                      {processing === inv.id ? 'Processing...' : 'Accept'}
                    </button>
                  </div>
                </div>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}
