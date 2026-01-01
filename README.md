# Base

Application starter template.

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
│   │   │   ├── auth/       # Google OAuth login/callback
│   │   │   ├── user/       # User persistence
│   │   │   ├── health/     # Health check endpoints
│   │   │   └── ping/       # Example: dual-DB writes (Postgres + DynamoDB)
│   │   ├── database/       # DB clients (postgres.go, dynamo.go, redis.go)
│   │   ├── session/        # Redis session management
│   │   ├── middleware/     # HTTP middleware (logging, auth, CORS)
│   │   ├── observability/  # CloudWatch metrics
│   │   └── router/         # Route mounting
│   └── pkg/response/       # Shared utilities
├── web/                    # Vite + React + TypeScript + Tailwind
│   └── src/
│       ├── components/     # Auth, layout, protected routes
│       ├── hooks/          # useAuth, useHealthCheck
│       └── router/         # React Router setup
├── database/
│   └── migrations/         # Goose SQL migrations (PostgreSQL)
├── infra/                  # Terraform (DynamoDB, CloudWatch)
│   └── environments/       # local.tfvars, prod.tfvars
└── docker-compose.yml
```

## Stack

| Component  | Purpose           | Management |
|------------|-------------------|------------|
| PostgreSQL | Users, relational data | Goose migrations (`database/migrations/`) |
| DynamoDB   | Time-series, NoSQL | Terraform (`infra/dynamodb.tf`) |
| Redis      | Sessions          | Docker (no schema) |

## Authentication

Google OAuth with Redis-backed sessions. Protected routes redirect to `/login`.

```
POST /auth/google/login     # Initiates OAuth flow
GET  /auth/google/callback  # OAuth callback, creates session
GET  /auth/me               # Current user (requires session)
POST /auth/logout           # Clears session
```

## Schema Changes

**PostgreSQL:** `make migrate-new name=<name>` then edit the generated file and `make migrate`

**DynamoDB:** Edit `infra/dynamodb.tf` then `make tf-plan && make tf-apply`

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
