import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useOrganization } from '../hooks/useOrganization'
import { MemberList } from '../components/organization/MemberList'
import { InviteForm } from '../components/organization/InviteForm'
import { OrganizationWithRole } from '../types/organization'

type Tab = 'general' | 'members' | 'invitations'

export function OrgSettings() {
  const { orgId } = useParams<{ orgId: string }>()
  const navigate = useNavigate()
  const { organizations, refetch } = useOrganization()
  const [org, setOrg] = useState<OrganizationWithRole | null>(null)
  const [activeTab, setActiveTab] = useState<Tab>('general')
  const [name, setName] = useState('')
  const [isSaving, setIsSaving] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)

  useEffect(() => {
    const foundOrg = organizations.find((o) => o.id === orgId)
    if (foundOrg) {
      setOrg(foundOrg)
      setName(foundOrg.name)
    }
  }, [organizations, orgId])

  if (!org) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-8">
        <div className="text-gray-400">Loading organization...</div>
      </div>
    )
  }

  const canManageMembers = org.role === 'owner' || org.role === 'admin'
  const canDeleteOrg = org.role === 'owner'

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim() || name === org.name) return

    setIsSaving(true)
    setError(null)
    setSuccess(null)

    try {
      const res = await fetch(`/api/organizations/${orgId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name }),
      })

      if (!res.ok) {
        const data = await res.json()
        throw new Error(data.message || 'Failed to update organization')
      }

      await refetch()
      setSuccess('Organization updated successfully')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update organization')
    } finally {
      setIsSaving(false)
    }
  }

  const handleLeave = async () => {
    if (!confirm('Are you sure you want to leave this organization?')) return

    try {
      const res = await fetch(`/api/organizations/${orgId}/leave`, {
        method: 'POST',
      })

      if (!res.ok) {
        const data = await res.json()
        throw new Error(data.message || 'Failed to leave organization')
      }

      await refetch()
      navigate('/organizations')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to leave organization')
    }
  }

  const handleDelete = async () => {
    if (!confirm('Are you sure you want to delete this organization? This action cannot be undone.')) {
      return
    }

    setIsDeleting(true)
    setError(null)

    try {
      const res = await fetch(`/api/organizations/${orgId}`, {
        method: 'DELETE',
      })

      if (!res.ok) {
        const data = await res.json()
        throw new Error(data.message || 'Failed to delete organization')
      }

      await refetch()
      navigate('/organizations')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete organization')
    } finally {
      setIsDeleting(false)
    }
  }

  const tabs: { id: Tab; label: string; show: boolean }[] = [
    { id: 'general', label: 'General', show: true },
    { id: 'members', label: 'Members', show: true },
    { id: 'invitations', label: 'Invitations', show: canManageMembers },
  ]

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">{org.name}</h1>
        <p className="text-gray-400">Organization Settings</p>
      </div>

      <div className="border-b border-gray-700 mb-6">
        <nav className="flex gap-6">
          {tabs.filter((t) => t.show).map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`pb-3 text-sm font-medium border-b-2 transition-colors ${
                activeTab === tab.id
                  ? 'border-blue-500 text-blue-400'
                  : 'border-transparent text-gray-400 hover:text-white'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {error && (
        <div className="mb-6 p-3 bg-red-900/50 border border-red-700 rounded-md text-red-300 text-sm">
          {error}
        </div>
      )}

      {success && (
        <div className="mb-6 p-3 bg-green-900/50 border border-green-700 rounded-md text-green-300 text-sm">
          {success}
        </div>
      )}

      {activeTab === 'general' && (
        <div className="space-y-6">
          <form onSubmit={handleSave} className="bg-gray-800 rounded-lg p-6">
            <h2 className="text-lg font-medium text-white mb-4">Organization Details</h2>
            <div className="mb-4">
              <label htmlFor="name" className="block text-sm font-medium text-gray-300 mb-2">
                Name
              </label>
              <input
                type="text"
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={!canManageMembers}
                className="w-full max-w-md px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:opacity-50 disabled:cursor-not-allowed"
              />
            </div>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Slug
              </label>
              <p className="text-gray-400">{org.slug}</p>
            </div>
            {canManageMembers && (
              <button
                type="submit"
                disabled={isSaving || name === org.name}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white text-sm font-medium rounded-md transition-colors"
              >
                {isSaving ? 'Saving...' : 'Save Changes'}
              </button>
            )}
          </form>

          <div className="bg-gray-800 rounded-lg p-6 border border-red-900/50">
            <h2 className="text-lg font-medium text-white mb-4">Danger Zone</h2>
            <div className="space-y-4">
              {!canDeleteOrg && (
                <div>
                  <button
                    onClick={handleLeave}
                    className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white text-sm font-medium rounded-md transition-colors"
                  >
                    Leave Organization
                  </button>
                  <p className="mt-1 text-sm text-gray-400">
                    You will lose access to this organization.
                  </p>
                </div>
              )}
              {canDeleteOrg && (
                <div>
                  <button
                    onClick={handleDelete}
                    disabled={isDeleting}
                    className="px-4 py-2 bg-red-600 hover:bg-red-700 disabled:bg-gray-600 text-white text-sm font-medium rounded-md transition-colors"
                  >
                    {isDeleting ? 'Deleting...' : 'Delete Organization'}
                  </button>
                  <p className="mt-1 text-sm text-gray-400">
                    This will permanently delete the organization and remove all members.
                  </p>
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {activeTab === 'members' && (
        <MemberList orgId={org.id} currentUserRole={org.role} />
      )}

      {activeTab === 'invitations' && canManageMembers && (
        <InviteForm orgId={org.id} />
      )}
    </div>
  )
}
