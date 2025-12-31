import { useEffect, useState } from 'react'

interface HealthResponse {
  status: string
  services: Record<string, string>
}

function App() {
  const [health, setHealth] = useState<HealthResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [lastChecked, setLastChecked] = useState<Date | null>(null)
  const [latency, setLatency] = useState<number | null>(null)
  const [secondsAgo, setSecondsAgo] = useState(0)

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

  useEffect(() => {
    fetchHealth()
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
        <h1 className="text-2xl font-bold mb-6">System Health</h1>

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
