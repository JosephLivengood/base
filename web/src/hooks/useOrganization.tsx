import { createContext, useContext, useEffect, useState, ReactNode, useCallback } from 'react'
import { OrganizationWithRole } from '../types/organization'
import { useAuth } from './useAuth'

interface OrganizationContextValue {
  organizations: OrganizationWithRole[]
  currentOrg: OrganizationWithRole | null
  isLoading: boolean
  switchOrg: (orgId: string) => Promise<void>
  refetch: () => Promise<void>
}

const OrganizationContext = createContext<OrganizationContextValue | null>(null)

const ACTIVE_ORG_KEY = 'activeOrgId'

export function OrganizationProvider({ children }: { children: ReactNode }) {
  const { isAuthenticated, isLoading: authLoading } = useAuth()
  const [organizations, setOrganizations] = useState<OrganizationWithRole[]>([])
  const [currentOrg, setCurrentOrg] = useState<OrganizationWithRole | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  const fetchOrganizations = useCallback(async () => {
    if (!isAuthenticated) {
      setOrganizations([])
      setCurrentOrg(null)
      setIsLoading(false)
      return
    }

    try {
      const res = await fetch('/api/organizations')
      const data = await res.json()
      const orgs: OrganizationWithRole[] = data.data || []
      setOrganizations(orgs)

      // Set current org from stored preference or first org
      const storedOrgId = localStorage.getItem(ACTIVE_ORG_KEY)
      const activeOrg = orgs.find((o) => o.id === storedOrgId) || orgs[0]
      if (activeOrg) {
        setCurrentOrg(activeOrg)
      }
    } catch {
      setOrganizations([])
      setCurrentOrg(null)
    } finally {
      setIsLoading(false)
    }
  }, [isAuthenticated])

  const switchOrg = useCallback(async (orgId: string) => {
    const org = organizations.find((o) => o.id === orgId)
    if (!org) return

    try {
      await fetch('/api/organizations/active', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ organization_id: orgId }),
      })
      localStorage.setItem(ACTIVE_ORG_KEY, orgId)
      setCurrentOrg(org)
    } catch (error) {
      console.error('Failed to switch organization:', error)
    }
  }, [organizations])

  useEffect(() => {
    if (!authLoading) {
      fetchOrganizations()
    }
  }, [authLoading, fetchOrganizations])

  return (
    <OrganizationContext.Provider
      value={{
        organizations,
        currentOrg,
        isLoading: isLoading || authLoading,
        switchOrg,
        refetch: fetchOrganizations,
      }}
    >
      {children}
    </OrganizationContext.Provider>
  )
}

export function useOrganization() {
  const context = useContext(OrganizationContext)
  if (!context) {
    throw new Error('useOrganization must be used within an OrganizationProvider')
  }
  return context
}
