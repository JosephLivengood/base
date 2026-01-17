package project

import "time"

type Project struct {
	ID             string     `json:"id" db:"id"`
	OrganizationID string     `json:"organization_id" db:"organization_id"`
	Name           string     `json:"name" db:"name"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type CreateProjectRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type UpdateProjectRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}
