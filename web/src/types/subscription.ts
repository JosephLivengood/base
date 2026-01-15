export type SubscriptionTier = 'free' | 'personal_plus'
export type SubscriptionStatus = 'active' | 'canceled' | 'past_due'

export interface Subscription {
  id: string
  organization_id: string
  tier: SubscriptionTier
  status: SubscriptionStatus
  stripe_customer_id?: string
  stripe_subscription_id?: string
  stripe_price_id?: string
  active_from: string
  active_until?: string
  created_at: string
}

export interface PricingTier {
  id: string
  name: string
  priceId: string
  price: number
  interval: 'month' | 'year'
}

export const PRICING: PricingTier[] = [
  {
    id: 'monthly',
    name: 'Personal Plus',
    priceId: 'price_1SpjQdQreiwkYmBAR7j1hsTv',
    price: 10,
    interval: 'month',
  },
  {
    id: 'yearly',
    name: 'Personal Plus',
    priceId: 'price_1SpjSlQreiwkYmBA4NmpIpiw',
    price: 95,
    interval: 'year',
  },
]

export interface CheckoutResponse {
  checkout_url: string
}

export interface PortalResponse {
  portal_url: string
}
