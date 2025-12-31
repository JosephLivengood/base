import { Link } from 'react-router-dom'
import { useAuth } from '../../hooks/useAuth'
import { GoogleLoginButton } from '../auth/GoogleLoginButton'
import { UserAvatar } from '../auth/UserAvatar'

export function Navigation() {
  const { user, isLoading } = useAuth()

  return (
    <nav className="border-b border-gray-800">
      <div className="max-w-4xl mx-auto px-4 py-4 flex items-center justify-between">
        <div className="flex items-center gap-6">
          <Link to="/" className="text-xl font-bold text-white">
            Base
          </Link>
          <div className="flex items-center gap-4">
            <Link
              to="/"
              className="text-sm text-gray-400 hover:text-white transition-colors"
            >
              Dashboard
            </Link>
            {user && (
              <Link
                to="/profile"
                className="text-sm text-gray-400 hover:text-white transition-colors"
              >
                Profile
              </Link>
            )}
          </div>
        </div>
        <div>
          {isLoading ? (
            <span className="text-sm text-gray-500">Loading...</span>
          ) : user ? (
            <UserAvatar user={user} />
          ) : (
            <GoogleLoginButton />
          )}
        </div>
      </div>
    </nav>
  )
}
