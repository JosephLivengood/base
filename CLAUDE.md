# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

```bash
make dev            # Start all services (Docker)
make dev-down       # Stop services
make dev-rebuild    # Rebuild and restart
make dev-logs       # View logs
make build          # Build API binary locally
make clean          # Clean build artifacts and docker volumes
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

**Existing domains**:
- `auth` - Google OAuth login/logout, session cookie management
- `user` - User persistence (Postgres)
- `health` - Health check endpoints (`/health/check`, `/health/ready`)
- `ping` - Example domain for testing DB operations

**Adding a new domain**:
1. Create `api/internal/domain/<name>/` with the files above
2. Mount in `api/internal/router/router.go` by creating repository → handler → calling `RegisterRoutes`

**Key packages**:
- `api/internal/database/` - Database clients (`postgres.go`, `dynamo.go`, `redis.go`)
- `api/internal/session/` - Redis-backed session store
- `api/internal/middleware/` - HTTP middleware (logging, CORS, auth, recovery, metrics)
- `api/pkg/response/` - JSON response helpers (`OK`, `Unauthorized`, `InternalError`, `JSON`)

**Authentication**:
- Google OAuth flow via `/api/auth/google` and `/api/auth/google/callback`
- Session stored in Redis, session ID in HTTP-only cookie
- Use `middleware.RequireAuth(sessionStore, userRepo)` to protect routes
- Access user/session in handlers via `middleware.GetUserFromContext(ctx)`

### Frontend (Vite + React + Tailwind)

React app with React Router in `web/`. API calls proxy through Vite dev server.

**Key structure**:
- `web/src/router/` - Route definitions with `ProtectedRoute` wrapper
- `web/src/hooks/useAuth.tsx` - Auth context provider and `useAuth()` hook
- `web/src/components/auth/` - Login button, user avatar, protected route
- `web/src/pages/` - Page components (Dashboard, Login, Profile, NotFound)

**Auth flow**: `AuthProvider` fetches `/auth/me` on load. Use `useAuth()` for `user`, `isAuthenticated`, `isLoading`.

### Database Schema

| Database   | Schema Management | Location |
|------------|-------------------|----------|
| PostgreSQL | Goose migrations  | `database/migrations/*.sql` |
| DynamoDB   | Terraform         | `infra/dynamodb.tf` |
| Redis      | N/A (session store) | — |
