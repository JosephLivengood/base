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
		INSERT INTO subscriptions (organization_id, tier)
		VALUES ($1, $2)
		RETURNING id, organization_id, tier, active_from, active_until, created_at
	`
	err := r.postgres.GetContext(ctx, &sub, query, orgID, tier)
	return &sub, err
}

func (r *Repository) CreateWithDates(ctx context.Context, orgID, tier string, activeFrom time.Time, activeUntil *time.Time) (*Subscription, error) {
	var sub Subscription
	query := `
		INSERT INTO subscriptions (organization_id, tier, active_from, active_until)
		VALUES ($1, $2, $3, $4)
		RETURNING id, organization_id, tier, active_from, active_until, created_at
	`
	err := r.postgres.GetContext(ctx, &sub, query, orgID, tier, activeFrom, activeUntil)
	return &sub, err
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Subscription, error) {
	var sub Subscription
	query := `
		SELECT id, organization_id, tier, active_from, active_until, created_at
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
		SELECT id, organization_id, tier, active_from, active_until, created_at
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
		SELECT id, organization_id, tier, active_from, active_until, created_at
		FROM subscriptions
		WHERE organization_id = $1
		  AND active_from <= NOW()
		  AND (active_until IS NULL OR active_until > NOW())
		ORDER BY created_at DESC
		LIMIT 1
	`
	err := r.postgres.GetContext(ctx, &sub, query, orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &sub, err
}
