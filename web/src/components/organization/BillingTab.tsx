import { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useSubscription } from '../../hooks/useSubscription'
import { PRICING } from '../../types/subscription'

interface BillingTabProps {
  orgId: string
}

export function BillingTab({ orgId }: BillingTabProps) {
  const { subscription, isLoading, error, createCheckout, openPortal, confirmCheckout, refetch } = useSubscription(orgId)
  const [searchParams, setSearchParams] = useSearchParams()
  const [checkoutError, setCheckoutError] = useState<string | null>(null)
  const [isProcessing, setIsProcessing] = useState(false)

  // Handle successful checkout redirect
  useEffect(() => {
    const sessionId = searchParams.get('session_id')
    if (sessionId && !isProcessing) {
      setIsProcessing(true)
      confirmCheckout(sessionId)
        .then(() => {
          // Remove session_id from URL
          searchParams.delete('session_id')
          setSearchParams(searchParams, { replace: true })
          refetch()
        })
        .catch((err) => {
          setCheckoutError(err instanceof Error ? err.message : 'Failed to confirm checkout')
          // Remove session_id from URL to prevent retry loop
          searchParams.delete('session_id')
          setSearchParams(searchParams, { replace: true })
        })
        .finally(() => {
          setIsProcessing(false)
        })
    }
  }, [searchParams, confirmCheckout, setSearchParams, refetch, isProcessing])

  const handleUpgrade = async (priceId: string) => {
    setCheckoutError(null)
    try {
      await createCheckout(priceId)
    } catch (err) {
      setCheckoutError(err instanceof Error ? err.message : 'Failed to start checkout')
    }
  }

  const handleManageSubscription = async () => {
    setCheckoutError(null)
    try {
      await openPortal()
    } catch (err) {
      setCheckoutError(err instanceof Error ? err.message : 'Failed to open billing portal')
    }
  }

  if (isLoading || isProcessing) {
    return (
      <div className="bg-gray-800 rounded-lg p-6">
        <div className="text-gray-400">Loading billing information...</div>
      </div>
    )
  }

  const hasActiveSubscription = subscription && subscription.tier !== 'free' && subscription.status === 'active'

  return (
    <div className="space-y-6">
      {(error || checkoutError) && (
        <div className="p-3 bg-red-900/50 border border-red-700 rounded-md text-red-300 text-sm">
          {error || checkoutError}
        </div>
      )}

      {/* Current Plan */}
      <div className="bg-gray-800 rounded-lg p-6">
        <h2 className="text-lg font-medium text-white mb-4">Current Plan</h2>
        <div className="flex items-center justify-between">
          <div>
            <p className="text-xl font-semibold text-white">
              {hasActiveSubscription ? 'Personal Plus' : 'Free'}
            </p>
            <p className="text-gray-400 text-sm mt-1">
              {hasActiveSubscription
                ? subscription.active_until
                  ? `Renews on ${new Date(subscription.active_until).toLocaleDateString()}`
                  : 'Active subscription'
                : 'Basic features included'}
            </p>
          </div>
          {hasActiveSubscription && (
            <button
              onClick={handleManageSubscription}
              className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white text-sm font-medium rounded-md transition-colors"
            >
              Manage Subscription
            </button>
          )}
        </div>
      </div>

      {/* Upgrade Options (only show if on free plan) */}
      {!hasActiveSubscription && (
        <div className="bg-gray-800 rounded-lg p-6">
          <h2 className="text-lg font-medium text-white mb-4">Upgrade to Personal Plus</h2>
          <p className="text-gray-400 mb-6">
            Unlock premium features and get more out of your organization.
          </p>
          <div className="grid md:grid-cols-2 gap-4">
            {PRICING.map((tier) => (
              <div
                key={tier.id}
                className="border border-gray-700 rounded-lg p-4 hover:border-blue-500 transition-colors"
              >
                <div className="flex justify-between items-start mb-3">
                  <div>
                    <h3 className="text-white font-medium">{tier.name}</h3>
                    <p className="text-sm text-gray-400">
                      Billed {tier.interval === 'month' ? 'monthly' : 'annually'}
                    </p>
                  </div>
                  <div className="text-right">
                    <p className="text-2xl font-bold text-white">${tier.price}</p>
                    <p className="text-sm text-gray-400">
                      /{tier.interval}
                    </p>
                  </div>
                </div>
                {tier.interval === 'year' && (
                  <div className="mb-3">
                    <span className="inline-block px-2 py-1 bg-green-900/50 text-green-400 text-xs font-medium rounded">
                      Save $25/year
                    </span>
                  </div>
                )}
                <button
                  onClick={() => handleUpgrade(tier.priceId)}
                  className="w-full px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-md transition-colors"
                >
                  Subscribe {tier.interval === 'month' ? 'Monthly' : 'Yearly'}
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Subscription Details (only show if subscribed) */}
      {hasActiveSubscription && subscription && (
        <div className="bg-gray-800 rounded-lg p-6">
          <h2 className="text-lg font-medium text-white mb-4">Subscription Details</h2>
          <dl className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <dt className="text-gray-400">Status</dt>
              <dd className="text-white capitalize">{subscription.status}</dd>
            </div>
            <div>
              <dt className="text-gray-400">Started</dt>
              <dd className="text-white">
                {new Date(subscription.active_from).toLocaleDateString()}
              </dd>
            </div>
            {subscription.active_until && (
              <div>
                <dt className="text-gray-400">Next Billing Date</dt>
                <dd className="text-white">
                  {new Date(subscription.active_until).toLocaleDateString()}
                </dd>
              </div>
            )}
          </dl>
        </div>
      )}
    </div>
  )
}
