import { useState, useRef, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useOrganization } from '../../hooks/useOrganization'

export function OrgSwitcher() {
  const { organizations, currentOrg, isLoading, switchOrg } = useOrganization()
  const [isOpen, setIsOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  if (isLoading) {
    return (
      <div className="text-sm text-gray-500 px-3 py-1.5">
        Loading...
      </div>
    )
  }

  if (!currentOrg) {
    return null
  }

  const handleSwitch = async (orgId: string) => {
    await switchOrg(orgId)
    setIsOpen(false)
  }

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 px-3 py-1.5 text-sm text-gray-300 hover:text-white bg-gray-800 hover:bg-gray-700 rounded-md transition-colors"
      >
        <span className="max-w-[150px] truncate">{currentOrg.name}</span>
        <svg
          className={`w-4 h-4 transition-transform ${isOpen ? 'rotate-180' : ''}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {isOpen && (
        <div className="absolute top-full left-0 mt-1 w-64 bg-gray-800 border border-gray-700 rounded-md shadow-lg z-50">
          <div className="py-1">
            <div className="px-3 py-2 text-xs font-medium text-gray-500 uppercase tracking-wider">
              Organizations
            </div>
            {organizations.map((org) => (
              <button
                key={org.id}
                onClick={() => handleSwitch(org.id)}
                className={`w-full text-left px-3 py-2 text-sm flex items-center justify-between hover:bg-gray-700 transition-colors ${
                  org.id === currentOrg.id ? 'text-white bg-gray-700' : 'text-gray-300'
                }`}
              >
                <span className="truncate">{org.name}</span>
                <span className="text-xs text-gray-500 capitalize">{org.role}</span>
              </button>
            ))}
          </div>
          <div className="border-t border-gray-700 py-1">
            <Link
              to="/organizations"
              onClick={() => setIsOpen(false)}
              className="block px-3 py-2 text-sm text-gray-400 hover:text-white hover:bg-gray-700 transition-colors"
            >
              Manage organizations
            </Link>
            <Link
              to={`/organizations/${currentOrg.id}/settings`}
              onClick={() => setIsOpen(false)}
              className="block px-3 py-2 text-sm text-gray-400 hover:text-white hover:bg-gray-700 transition-colors"
            >
              Organization settings
            </Link>
          </div>
        </div>
      )}
    </div>
  )
}
