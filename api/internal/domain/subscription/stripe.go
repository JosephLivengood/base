package subscription

import (
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/customer"
)

type StripeService struct {
	secretKey     string
	webhookSecret string
	environment   string
}

func NewStripeService(secretKey, webhookSecret, environment string) *StripeService {
	stripe.Key = secretKey
	return &StripeService{
		secretKey:     secretKey,
		webhookSecret: webhookSecret,
		environment:   environment,
	}
}

// CreateCheckoutSession creates a Stripe Checkout session for subscription upgrade
func (s *StripeService) CreateCheckoutSession(customerID, priceID, successURL, cancelURL string) (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
	}

	return checkoutsession.New(params)
}

// CreateBillingPortalSession creates a Stripe Billing Portal session
func (s *StripeService) CreateBillingPortalSession(customerID, returnURL string) (*stripe.BillingPortalSession, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	}

	return session.New(params)
}

// GetOrCreateCustomer gets an existing customer or creates a new one
func (s *StripeService) GetOrCreateCustomer(orgID, orgName, email string) (*stripe.Customer, error) {
	// Search for existing customer by metadata
	searchParams := &stripe.CustomerSearchParams{
		SearchParams: stripe.SearchParams{
			Query: "metadata['organization_id']:'" + orgID + "'",
		},
	}

	iter := customer.Search(searchParams)
	for iter.Next() {
		return iter.Customer(), nil
	}

	// Create new customer if not found
	params := &stripe.CustomerParams{
		Name:  stripe.String(orgName),
		Email: stripe.String(email),
		Metadata: map[string]string{
			"organization_id": orgID,
		},
	}

	return customer.New(params)
}

// GetCheckoutSession retrieves a checkout session by ID
func (s *StripeService) GetCheckoutSession(sessionID string) (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionParams{}
	params.AddExpand("subscription")
	params.AddExpand("customer")

	return checkoutsession.Get(sessionID, params)
}

// WebhookSecret returns the webhook secret for signature verification
func (s *StripeService) WebhookSecret() string {
	return s.webhookSecret
}

// IsDevelopment returns true if running in development mode
func (s *StripeService) IsDevelopment() bool {
	return s.environment == "development"
}
