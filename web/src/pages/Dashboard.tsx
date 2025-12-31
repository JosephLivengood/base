import { useHealthCheck } from '../hooks/useHealthCheck'
import { OverallStatus } from '../components/dashboard/OverallStatus'
import { ServiceStatus } from '../components/dashboard/ServiceStatus'

export function Dashboard() {
  const { health, error, isLoading, lastChecked, latency, secondsAgo } =
    useHealthCheck()

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-gray-400">Loading...</p>
      </div>
    )
  }

  return (
    <div className="max-w-md mx-auto">
      <h1 className="text-2xl font-bold mb-6">System Health</h1>

      {error ? (
        <div className="bg-red-900/50 border border-red-500 rounded-lg p-4">
          <p className="text-red-400">{error}</p>
        </div>
      ) : health ? (
        <div className="space-y-4">
          <OverallStatus status={health.status} />

          <div className="bg-gray-800 rounded-lg divide-y divide-gray-700">
            {Object.entries(health.services).map(([name, status]) => (
              <ServiceStatus key={name} name={name} status={status} />
            ))}
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
  )
}
