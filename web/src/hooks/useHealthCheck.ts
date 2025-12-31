import { useEffect, useState } from 'react'
import { HealthResponse } from '../types/health'

interface UseHealthCheckReturn {
  health: HealthResponse | null
  error: string | null
  isLoading: boolean
  lastChecked: Date | null
  latency: number | null
  secondsAgo: number
  refetch: () => Promise<void>
}

export function useHealthCheck(intervalMs = 30000): UseHealthCheckReturn {
  const [health, setHealth] = useState<HealthResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
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
    } catch {
      setLatency(null)
      setError('Failed to connect to API')
      setHealth(null)
    } finally {
      setLastChecked(new Date())
      setSecondsAgo(0)
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchHealth()
    const interval = setInterval(fetchHealth, intervalMs)
    return () => clearInterval(interval)
  }, [intervalMs])

  useEffect(() => {
    const timer = setInterval(() => {
      if (lastChecked) {
        setSecondsAgo(Math.floor((Date.now() - lastChecked.getTime()) / 1000))
      }
    }, 1000)
    return () => clearInterval(timer)
  }, [lastChecked])

  return {
    health,
    error,
    isLoading,
    lastChecked,
    latency,
    secondsAgo,
    refetch: fetchHealth,
  }
}
