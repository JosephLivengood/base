package subscription

import "time"

// Tier constants
const (
	TierFree         = "free"
	TierPersonalPlus = "personal_plus"
)

// Status constants
const (
	StatusActive   = "active"
	StatusCanceled = "canceled"
	StatusPastDue  = "past_due"
)

// Price IDs
const (
	PricePersonalPlusMonthly = "price_1SpjQdQreiwkYmBAR7j1hsTv"
	PricePersonalPlusYearly  = "price_1SpjSlQreiwkYmBA4NmpIpiw"
)

type Subscription struct {
	ID                   string     `json:"id" db:"id"`
	OrganizationID       string     `json:"organization_id" db:"organization_id"`
	Tier                 string     `json:"tier" db:"tier"`
	Status               string     `json:"status" db:"status"`
	StripeCustomerID     *string    `json:"stripe_customer_id,omitempty" db:"stripe_customer_id"`
	StripeSubscriptionID *string    `json:"stripe_subscription_id,omitempty" db:"stripe_subscription_id"`
	StripePriceID        *string    `json:"stripe_price_id,omitempty" db:"stripe_price_id"`
	ActiveFrom           time.Time  `json:"active_from" db:"active_from"`
	ActiveUntil          *time.Time `json:"active_until,omitempty" db:"active_until"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
}

// CheckoutRequest is the request body for creating a checkout session
type CheckoutRequest struct {
	PriceID    string `json:"price_id"`
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
}

// CheckoutResponse is the response body for creating a checkout session
type CheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
}

// PortalRequest is the request body for creating a billing portal session
type PortalRequest struct {
	ReturnURL string `json:"return_url"`
}

// PortalResponse is the response body for creating a billing portal session
type PortalResponse struct {
	PortalURL string `json:"portal_url"`
}

// ConfirmRequest is the request body for confirming a checkout session (dev only)
type ConfirmRequest struct {
	SessionID string `json:"session_id"`
}
