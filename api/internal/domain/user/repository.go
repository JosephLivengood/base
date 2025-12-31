package user

import (
	"context"
	"database/sql"
	"errors"

	"base/api/internal/database"
)

var ErrNotFound = errors.New("user not found")

type Repository struct {
	postgres *database.PostgresDB
}

func NewRepository(postgres *database.PostgresDB) *Repository {
	return &Repository{postgres: postgres}
}

func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	var user User
	query := `SELECT id, email, name, picture, google_id, created_at, updated_at FROM users WHERE id = $1`
	err := r.postgres.GetContext(ctx, &user, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &user, err
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	query := `SELECT id, email, name, picture, google_id, created_at, updated_at FROM users WHERE email = $1`
	err := r.postgres.GetContext(ctx, &user, query, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &user, err
}

func (r *Repository) GetByGoogleID(ctx context.Context, googleID string) (*User, error) {
	var user User
	query := `SELECT id, email, name, picture, google_id, created_at, updated_at FROM users WHERE google_id = $1`
	err := r.postgres.GetContext(ctx, &user, query, googleID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &user, err
}

func (r *Repository) Upsert(ctx context.Context, email, name, picture, googleID string) (*User, error) {
	var user User
	query := `
		INSERT INTO users (email, name, picture, google_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) DO UPDATE SET
			name = EXCLUDED.name,
			picture = EXCLUDED.picture,
			google_id = COALESCE(users.google_id, EXCLUDED.google_id),
			updated_at = NOW()
		RETURNING id, email, name, picture, google_id, created_at, updated_at
	`
	err := r.postgres.GetContext(ctx, &user, query, email, name, picture, googleID)
	return &user, err
}
