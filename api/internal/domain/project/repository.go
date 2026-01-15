package project

import (
	"context"
	"database/sql"
	"errors"

	"base/api/internal/database"
)

var ErrNotFound = errors.New("project not found")

type Repository struct {
	postgres *database.PostgresDB
}

func NewRepository(postgres *database.PostgresDB) *Repository {
	return &Repository{postgres: postgres}
}

func (r *Repository) Create(ctx context.Context, orgID, name string) (*Project, error) {
	var project Project
	query := `
		INSERT INTO projects (organization_id, name)
		VALUES ($1, $2)
		RETURNING id, organization_id, name, created_at, deleted_at
	`
	err := r.postgres.GetContext(ctx, &project, query, orgID, name)
	return &project, err
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Project, error) {
	var project Project
	query := `
		SELECT id, organization_id, name, created_at, deleted_at
		FROM projects
		WHERE id = $1 AND deleted_at IS NULL
	`
	err := r.postgres.GetContext(ctx, &project, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &project, err
}

func (r *Repository) GetByOrganization(ctx context.Context, orgID string) ([]Project, error) {
	var projects []Project
	query := `
		SELECT id, organization_id, name, created_at, deleted_at
		FROM projects
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`
	err := r.postgres.SelectContext(ctx, &projects, query, orgID)
	return projects, err
}

func (r *Repository) Update(ctx context.Context, id, name string) (*Project, error) {
	var project Project
	query := `
		UPDATE projects
		SET name = $2
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, organization_id, name, created_at, deleted_at
	`
	err := r.postgres.GetContext(ctx, &project, query, id, name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &project, err
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE projects
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	result, err := r.postgres.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
