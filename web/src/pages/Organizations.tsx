import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useOrganization } from '../hooks/useOrganization'

export function Organizations() {
  const { organizations, isLoading, refetch } = useOrganization()
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [newOrgName, setNewOrgName] = useState('')
  const [isCreating, setIsCreating] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newOrgName.trim()) return

    setIsCreating(true)
    setError(null)

    try {
      const res = await fetch('/api/organizations', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: newOrgName }),
      })

      if (!res.ok) {
        const data = await res.json()
        throw new Error(data.message || 'Failed to create organization')
      }

      await refetch()
      setShowCreateModal(false)
      setNewOrgName('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create organization')
    } finally {
      setIsCreating(false)
    }
  }

  if (isLoading) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-8">
        <div className="text-gray-400">Loading organizations...</div>
      </div>
    )
  }

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-white">Organizations</h1>
        <button
          onClick={() => setShowCreateModal(true)}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-md transition-colors"
        >
          Create Organization
        </button>
      </div>

      <div className="bg-gray-800 rounded-lg overflow-hidden">
        {organizations.length === 0 ? (
          <div className="px-6 py-12 text-center text-gray-400">
            No organizations yet. Create your first one!
          </div>
        ) : (
          <ul className="divide-y divide-gray-700">
            {organizations.map((org) => (
              <li key={org.id}>
                <Link
                  to={`/organizations/${org.id}/settings`}
                  className="block px-6 py-4 hover:bg-gray-700 transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <h3 className="text-white font-medium">{org.name}</h3>
                      <p className="text-sm text-gray-400">{org.slug}</p>
                    </div>
                    <span className="px-2 py-1 text-xs font-medium bg-gray-700 text-gray-300 rounded capitalize">
                      {org.role}
                    </span>
                  </div>
                </Link>
              </li>
            ))}
          </ul>
        )}
      </div>

      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-lg p-6 w-full max-w-md mx-4">
            <h2 className="text-xl font-bold text-white mb-4">Create Organization</h2>
            <form onSubmit={handleCreate}>
              <div className="mb-4">
                <label htmlFor="orgName" className="block text-sm font-medium text-gray-300 mb-2">
                  Organization Name
                </label>
                <input
                  type="text"
                  id="orgName"
                  value={newOrgName}
                  onChange={(e) => setNewOrgName(e.target.value)}
                  className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="Enter organization name"
                  autoFocus
                />
              </div>
              {error && (
                <p className="mb-4 text-sm text-red-400">{error}</p>
              )}
              <div className="flex gap-3 justify-end">
                <button
                  type="button"
                  onClick={() => {
                    setShowCreateModal(false)
                    setNewOrgName('')
                    setError(null)
                  }}
                  className="px-4 py-2 text-sm font-medium text-gray-300 hover:text-white transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isCreating || !newOrgName.trim()}
                  className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white text-sm font-medium rounded-md transition-colors"
                >
                  {isCreating ? 'Creating...' : 'Create'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
