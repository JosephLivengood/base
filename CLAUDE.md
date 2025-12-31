# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

```bash
make dev            # Start all services (Docker)
make dev-down       # Stop services
make dev-rebuild    # Rebuild and restart
make dev-logs       # View logs
```

- Frontend: http://localhost:5173
- API: http://localhost:8080

### Database Migrations (PostgreSQL)

```bash
make migrate                          # Run pending migrations
make migrate-new name=<name>          # Create new migration
make migrate-down                     # Rollback last migration
make migrate-status                   # Show migration status
```

### Terraform (DynamoDB)

```bash
make tf-init        # One-time init
make tf-plan        # Review local changes
make tf-apply       # Apply local changes
make tf-apply-prod  # Apply production changes
```

## Architecture

### Backend (Go + Chi)

The API uses a domain-driven structure where each feature is a self-contained module.

**Domain module pattern** (`api/internal/domain/<feature>/`):
- `models.go` - Data structures and types
- `repository.go` - Database operations (Postgres + DynamoDB)
- `handler.go` - HTTP handlers
- `routes.go` - Route registration via `RegisterRoutes(r chi.Router, h *Handler)`

**Adding a new domain**:
1. Create `api/internal/domain/<name>/` with the 4 files above
2. Mount in `api/internal/router/router.go` by creating repository → handler → calling `RegisterRoutes`

**Key packages**:
- `api/internal/database/` - Database clients (`postgres.go`, `dynamo.go`)
- `api/internal/middleware/` - HTTP middleware (logging, CORS, auth)
- `api/pkg/response/` - JSON response helpers (`OK`, `InternalError`, `JSON`)

### Frontend (Vite + React + Tailwind)

Standard Vite React setup in `web/`. API calls proxy through Vite dev server (`/api` → `http://api:8080`).

### Database Schema

| Database   | Schema Management | Location |
|------------|-------------------|----------|
| PostgreSQL | Goose migrations  | `database/migrations/*.sql` |
| DynamoDB   | Terraform         | `infra/dynamodb.tf` |
