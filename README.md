# Base

SaaS starter template.

## Quickstart

Prerequisites: Docker, Terraform

```bash
make dev          # Start all services
make tf-init      # One-time Terraform init
make tf-apply     # Create DynamoDB tables
```

- Frontend: http://localhost:5173
- API: http://localhost:8080

## Project Structure

```
├── api/                    # Go backend (Chi router, sqlx)
│   ├── internal/
│   │   ├── domain/         # Feature modules (add new features here)
│   │   │   ├── health/     # Health check endpoints
│   │   │   └── ping/       # Example: handler, repository, models, routes
│   │   ├── database/       # DB clients (postgres.go, dynamo.go)
│   │   ├── middleware/     # HTTP middleware
│   │   └── router/         # Route mounting
│   └── pkg/response/       # Shared utilities
├── web/                    # Vite + React + Tailwind frontend
├── database/
│   └── migrations/         # Goose SQL migrations (PostgreSQL)
├── infra/                  # Terraform (DynamoDB tables)
└── docker-compose.yml
```

## Database Patterns

| Database   | Schema Management | Location |
|------------|-------------------|----------|
| PostgreSQL | Goose migrations  | `database/migrations/*.sql` |
| DynamoDB   | Terraform         | `infra/dynamodb.tf` |

### PostgreSQL: Add a migration
```bash
make migrate-new name=create_users_table
# Edit database/migrations/XXXXX_create_users_table.sql
make migrate
```

### DynamoDB: Add a table
Edit `infra/dynamodb.tf`, then:
```bash
make tf-plan    # Review
make tf-apply   # Apply
```

## Adding a New API Domain

1. Create `api/internal/domain/<name>/`
2. Add files: `models.go`, `repository.go`, `handler.go`, `routes.go`
3. Mount in `api/internal/router/router.go`

## Commands

```bash
make dev            # Start all services
make dev-down       # Stop services
make dev-rebuild    # Rebuild and restart
make migrate        # Run pending migrations
make tf-apply       # Apply Terraform (local)
make tf-apply-prod  # Apply Terraform (production)
```
