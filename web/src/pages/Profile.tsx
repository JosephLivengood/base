import { useAuth } from '../hooks/useAuth'

export function Profile() {
  const { user } = useAuth()

  if (!user) {
    return null
  }

  return (
    <div className="max-w-md mx-auto">
      <h1 className="text-2xl font-bold mb-6">Profile</h1>

      <div className="bg-gray-800 rounded-lg p-6">
        <div className="flex items-center gap-4 mb-6">
          <img
            src={user.picture}
            alt={user.name}
            className="w-16 h-16 rounded-full"
          />
          <div>
            <h2 className="text-xl font-semibold">{user.name}</h2>
            <p className="text-gray-400">{user.email}</p>
          </div>
        </div>

        <div className="space-y-3 text-sm">
          <div className="flex justify-between">
            <span className="text-gray-400">User ID</span>
            <span className="font-mono">{user.id}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-gray-400">Member since</span>
            <span>{new Date(user.created_at).toLocaleDateString()}</span>
          </div>
        </div>
      </div>
    </div>
  )
}
