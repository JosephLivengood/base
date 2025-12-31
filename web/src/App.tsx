import { useEffect, useState } from 'react'

interface HealthResponse {
  status: string
  services: Record<string, string>
}

interface GoogleUser {
  id: string
  email: string
  name: string
  picture: string
}

function App() {
  const [health, setHealth] = useState<HealthResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [lastChecked, setLastChecked] = useState<Date | null>(null)
  const [latency, setLatency] = useState<number | null>(null)
  const [secondsAgo, setSecondsAgo] = useState(0)
  const [user, setUser] = useState<GoogleUser | null>(null)

  const fetchHealth = async () => {
    const start = performance.now()
    try {
      const res = await fetch('/health/ready')
      const data = await res.json()
      setLatency(Math.round(performance.now() - start))
      setHealth(data)
      setError(null)
      setLastChecked(new Date())
      setSecondsAgo(0)
    } catch (err) {
      setLatency(null)
      setError('Failed to connect to API')
      setHealth(null)
      setLastChecked(new Date())
      setSecondsAgo(0)
    } finally {
      setLoading(false)
    }
  }

  const fetchUser = async () => {
    try {
      const res = await fetch('/auth/me')
      const data = await res.json()
      setUser(data.data)
    } catch {
      setUser(null)
    }
  }

  useEffect(() => {
    fetchHealth()
    fetchUser()
    const interval = setInterval(fetchHealth, 30000)
    return () => clearInterval(interval)
  }, [])

  useEffect(() => {
    const timer = setInterval(() => {
      if (lastChecked) {
        setSecondsAgo(Math.floor((Date.now() - lastChecked.getTime()) / 1000))
      }
    }, 1000)
    return () => clearInterval(timer)
  }, [lastChecked])

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-900 flex items-center justify-center">
        <p className="text-gray-400">Loading...</p>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-900 text-white p-8">
      <div className="max-w-md mx-auto">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-2xl font-bold">System Health</h1>
          {user ? (
            <div className="flex items-center gap-3">
              <img
                src={user.picture}
                alt={user.name}
                className="w-8 h-8 rounded-full"
              />
              <span className="text-sm text-gray-300">{user.name}</span>
              <a
                href="/auth/logout"
                className="text-sm text-gray-400 hover:text-white"
              >
                Sign out
              </a>
            </div>
          ) : (
            <a
              href="/auth/google"
              className="flex items-center gap-2 bg-white text-gray-900 px-4 py-2 rounded-lg font-medium hover:bg-gray-100 transition-colors"
            >
              <svg className="w-5 h-5" viewBox="0 0 24 24">
                <path
                  fill="#4285F4"
                  d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                />
                <path
                  fill="#34A853"
                  d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                />
                <path
                  fill="#FBBC05"
                  d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
                />
                <path
                  fill="#EA4335"
                  d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                />
              </svg>
              Sign in with Google
            </a>
          )}
        </div>

        {error ? (
          <div className="bg-red-900/50 border border-red-500 rounded-lg p-4">
            <p className="text-red-400">{error}</p>
          </div>
        ) : health ? (
          <div className="space-y-4">
            <div
              className={`rounded-lg p-4 ${
                health.status === 'ok'
                  ? 'bg-green-900/50 border border-green-500'
                  : 'bg-yellow-900/50 border border-yellow-500'
              }`}
            >
              <p className="text-sm text-gray-400">Overall Status</p>
              <p
                className={`text-xl font-semibold ${
                  health.status === 'ok' ? 'text-green-400' : 'text-yellow-400'
                }`}
              >
                {health.status.toUpperCase()}
              </p>
            </div>

            <div className="bg-gray-800 rounded-lg divide-y divide-gray-700">
              {Object.entries(health.services).map(([name, status]) => {
                const isHealthy = status === 'healthy'
                return (
                  <div key={name} className="p-4 flex items-center justify-between">
                    <span className="font-medium">{name}</span>
                    <div className="flex items-center gap-2">
                      <span
                        className={`w-2 h-2 rounded-full ${
                          isHealthy ? 'bg-green-500' : 'bg-red-500'
                        }`}
                      />
                      <span
                        className={`text-sm ${
                          isHealthy ? 'text-green-400' : 'text-red-400'
                        }`}
                      >
                        {isHealthy ? 'healthy' : status}
                      </span>
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        ) : null}

        {lastChecked && (
          <div className="mt-6 text-center text-sm text-gray-500">
            Last checked {secondsAgo}s ago
            {latency !== null && <span> · {latency}ms</span>}
          </div>
        )}
      </div>
    </div>
  )
}

export default App
