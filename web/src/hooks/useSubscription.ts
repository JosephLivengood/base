import { useState, useCallback, useEffect } from 'react'
import { Subscription, CheckoutResponse, PortalResponse } from '../types/subscription'

interface UseSubscriptionResult {
  subscription: Subscription | null
  isLoading: boolean
  error: string | null
  createCheckout: (priceId: string) => Promise<void>
  openPortal: () => Promise<void>
  confirmCheckout: (sessionId: string) => Promise<void>
  refetch: () => Promise<void>
}

export function useSubscription(orgId: string | undefined): UseSubscriptionResult {
  const [subscription, setSubscription] = useState<Subscription | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchSubscription = useCallback(async () => {
    if (!orgId) {
      setSubscription(null)
      setIsLoading(false)
      return
    }

    setIsLoading(true)
    setError(null)

    try {
      const res = await fetch(`/api/organizations/${orgId}/subscription`)
      if (!res.ok) {
        throw new Error('Failed to fetch subscription')
      }
      const data = await res.json()
      setSubscription(data.data || null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch subscription')
      setSubscription(null)
    } finally {
      setIsLoading(false)
    }
  }, [orgId])

  const createCheckout = useCallback(async (priceId: string) => {
    if (!orgId) return

    const successUrl = `${window.location.origin}/organizations/${orgId}/settings?tab=billing&session_id={CHECKOUT_SESSION_ID}`
    const cancelUrl = `${window.location.origin}/organizations/${orgId}/settings?tab=billing`

    const res = await fetch(`/api/organizations/${orgId}/subscription/checkout`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        price_id: priceId,
        success_url: successUrl,
        cancel_url: cancelUrl,
      }),
    })

    if (!res.ok) {
      const data = await res.json()
      throw new Error(data.message || 'Failed to create checkout session')
    }

    const data: { data: CheckoutResponse } = await res.json()

    // Redirect to Stripe Checkout
    window.location.href = data.data.checkout_url
  }, [orgId])

  const openPortal = useCallback(async () => {
    if (!orgId) return

    const returnUrl = `${window.location.origin}/organizations/${orgId}/settings?tab=billing`

    const res = await fetch(`/api/organizations/${orgId}/subscription/portal`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        return_url: returnUrl,
      }),
    })

    if (!res.ok) {
      const data = await res.json()
      throw new Error(data.message || 'Failed to create portal session')
    }

    const data: { data: PortalResponse } = await res.json()

    // Redirect to Stripe Billing Portal
    window.location.href = data.data.portal_url
  }, [orgId])

  const confirmCheckout = useCallback(async (sessionId: string) => {
    if (!orgId) return

    const res = await fetch(`/api/organizations/${orgId}/subscription/confirm`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        session_id: sessionId,
      }),
    })

    if (!res.ok) {
      const data = await res.json()
      throw new Error(data.message || 'Failed to confirm checkout')
    }

    const data = await res.json()
    setSubscription(data.data || null)
  }, [orgId])

  useEffect(() => {
    fetchSubscription()
  }, [fetchSubscription])

  return {
    subscription,
    isLoading,
    error,
    createCheckout,
    openPortal,
    confirmCheckout,
    refetch: fetchSubscription,
  }
}
