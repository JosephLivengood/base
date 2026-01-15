package subscription

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"base/api/internal/database"
)

var ErrNotFound = errors.New("subscription not found")

type Repository struct {
	postgres *database.PostgresDB
}

func NewRepository(postgres *database.PostgresDB) *Repository {
	return &Repository{postgres: postgres}
}

func (r *Repository) Create(ctx context.Context, orgID, tier string) (*Subscription, error) {
	var sub Subscription
	query := `
		INSERT INTO subscriptions (organization_id, tier, status)
		VALUES ($1, $2, $3)
		RETURNING id, organization_id, tier, status, stripe_customer_id, stripe_subscription_id, stripe_price_id, active_from, active_until, created_at
	`
	err := r.postgres.GetContext(ctx, &sub, query, orgID, tier, StatusActive)
	return &sub, err
}

func (r *Repository) CreateWithDates(ctx context.Context, orgID, tier string, activeFrom time.Time, activeUntil *time.Time) (*Subscription, error) {
	var sub Subscription
	query := `
		INSERT INTO subscriptions (organization_id, tier, status, active_from, active_until)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, organization_id, tier, status, stripe_customer_id, stripe_subscription_id, stripe_price_id, active_from, active_until, created_at
	`
	err := r.postgres.GetContext(ctx, &sub, query, orgID, tier, StatusActive, activeFrom, activeUntil)
	return &sub, err
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Subscription, error) {
	var sub Subscription
	query := `
		SELECT id, organization_id, tier, status, stripe_customer_id, stripe_subscription_id, stripe_price_id, active_from, active_until, created_at
		FROM subscriptions
		WHERE id = $1
	`
	err := r.postgres.GetContext(ctx, &sub, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &sub, err
}

func (r *Repository) GetByOrganization(ctx context.Context, orgID string) ([]Subscription, error) {
	var subs []Subscription
	query := `
		SELECT id, organization_id, tier, status, stripe_customer_id, stripe_subscription_id, stripe_price_id, active_from, active_until, created_at
		FROM subscriptions
		WHERE organization_id = $1
		ORDER BY created_at DESC
	`
	err := r.postgres.SelectContext(ctx, &subs, query, orgID)
	return subs, err
}

func (r *Repository) GetActiveByOrganization(ctx context.Context, orgID string) (*Subscription, error) {
	var sub Subscription
	query := `
		SELECT id, organization_id, tier, status, stripe_customer_id, stripe_subscription_id, stripe_price_id, active_from, active_until, created_at
		FROM subscriptions
		WHERE organization_id = $1
		  AND active_from <= NOW()
		  AND (active_until IS NULL OR active_until > NOW())
		  AND status = $2
		ORDER BY created_at DESC
		LIMIT 1
	`
	err := r.postgres.GetContext(ctx, &sub, query, orgID, StatusActive)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &sub, err
}

// GetByStripeSubscriptionID finds a subscription by its Stripe subscription ID
func (r *Repository) GetByStripeSubscriptionID(ctx context.Context, stripeSubID string) (*Subscription, error) {
	var sub Subscription
	query := `
		SELECT id, organization_id, tier, status, stripe_customer_id, stripe_subscription_id, stripe_price_id, active_from, active_until, created_at
		FROM subscriptions
		WHERE stripe_subscription_id = $1
	`
	err := r.postgres.GetContext(ctx, &sub, query, stripeSubID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &sub, err
}

// GetByStripeCustomerID finds a subscription by its Stripe customer ID
func (r *Repository) GetByStripeCustomerID(ctx context.Context, customerID string) (*Subscription, error) {
	var sub Subscription
	query := `
		SELECT id, organization_id, tier, status, stripe_customer_id, stripe_subscription_id, stripe_price_id, active_from, active_until, created_at
		FROM subscriptions
		WHERE stripe_customer_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	err := r.postgres.GetContext(ctx, &sub, query, customerID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &sub, err
}

// CreateFromStripe creates a new subscription with Stripe data
func (r *Repository) CreateFromStripe(ctx context.Context, orgID, tier, status string,
	stripeCustomerID, stripeSubscriptionID, stripePriceID *string,
	activeFrom time.Time, activeUntil *time.Time) (*Subscription, error) {

	var sub Subscription
	query := `
		INSERT INTO subscriptions (organization_id, tier, status, stripe_customer_id, stripe_subscription_id, stripe_price_id, active_from, active_until)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, organization_id, tier, status, stripe_customer_id, stripe_subscription_id, stripe_price_id, active_from, active_until, created_at
	`
	err := r.postgres.GetContext(ctx, &sub, query, orgID, tier, status, stripeCustomerID, stripeSubscriptionID, stripePriceID, activeFrom, activeUntil)
	return &sub, err
}

// UpdateStatus updates the status of a subscription
func (r *Repository) UpdateStatus(ctx context.Context, id, status string) (*Subscription, error) {
	var sub Subscription
	query := `
		UPDATE subscriptions
		SET status = $2
		WHERE id = $1
		RETURNING id, organization_id, tier, status, stripe_customer_id, stripe_subscription_id, stripe_price_id, active_from, active_until, created_at
	`
	err := r.postgres.GetContext(ctx, &sub, query, id, status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &sub, err
}

// UpdateStripeInfo updates the Stripe-related fields of a subscription
func (r *Repository) UpdateStripeInfo(ctx context.Context, id string,
	stripeCustomerID, stripeSubscriptionID, stripePriceID *string,
	tier, status string, activeUntil *time.Time) (*Subscription, error) {

	var sub Subscription
	query := `
		UPDATE subscriptions
		SET stripe_customer_id = $2,
		    stripe_subscription_id = $3,
		    stripe_price_id = $4,
		    tier = $5,
		    status = $6,
		    active_until = $7
		WHERE id = $1
		RETURNING id, organization_id, tier, status, stripe_customer_id, stripe_subscription_id, stripe_price_id, active_from, active_until, created_at
	`
	err := r.postgres.GetContext(ctx, &sub, query, id, stripeCustomerID, stripeSubscriptionID, stripePriceID, tier, status, activeUntil)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &sub, err
}
