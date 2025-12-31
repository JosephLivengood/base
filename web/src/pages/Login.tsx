import { Navigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { GoogleLoginButton } from '../components/auth/GoogleLoginButton'

export function Login() {
  const { user, isLoading } = useAuth()

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-900 flex items-center justify-center">
        <p className="text-gray-400">Loading...</p>
      </div>
    )
  }

  if (user) {
    return <Navigate to="/" replace />
  }

  return (
    <div className="min-h-screen bg-gray-900 flex items-center justify-center">
      <div className="text-center">
        <h1 className="text-3xl font-bold text-white mb-8">Welcome</h1>
        <GoogleLoginButton />
      </div>
    </div>
  )
}
