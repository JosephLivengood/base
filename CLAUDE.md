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
- `organization` - Multi-tenant organizations, members, invitations
- `project` - Projects belonging to organizations (soft delete)
- `subscription` - Stripe subscriptions and billing
- `health` - Health check endpoints (`/health/check`, `/health/ready`)
- `ping` - Example domain for testing DB operations

**Adding a new domain**:
1. Create `api/internal/domain/<name>/` with the files above
2. Mount in `api/internal/router/router.go` by creating repository → handler → calling `RegisterRoutes`

**Key packages**:
- `api/internal/database/` - Database clients (`postgres.go`, `dynamo.go`, `redis.go`)
- `api/internal/session/` - Redis-backed session store (includes `ActiveOrgID`)
- `api/internal/middleware/` - HTTP middleware (logging, CORS, auth, recovery, metrics)
- `api/pkg/response/` - JSON response helpers (`OK`, `Unauthorized`, `InternalError`, `JSON`)

**Authentication**:
- Google OAuth flow via `/api/auth/google` and `/api/auth/google/callback`
- Session stored in Redis, session ID in HTTP-only cookie
- Use `middleware.RequireAuth(sessionStore, userRepo)` to protect routes
- Access user/session in handlers via `middleware.GetUserFromContext(ctx)`
- New users automatically get a "Personal" organization on signup

**Organizations**:
- Multi-tenant system: users belong to organizations with roles (`owner`, `admin`, `member`)
- Organization context tracked in session via `ActiveOrgID`
- Invitation system for adding members by email
- Routes nested under `/api/organizations/{orgID}/...`

**Subscriptions (Stripe)**:
- Tiers: `free`, `personal_plus`
- Stripe Checkout for payment, Billing Portal for management
- Local dev uses optimistic confirm endpoint (no webhooks needed)
- Production uses Stripe webhooks at `/api/stripe/webhook`
- Environment: `STRIPE_SECRET_KEY`, `STRIPE_PUBLISHABLE_KEY`, `STRIPE_WEBHOOK_SECRET`

### Frontend (Vite + React + Tailwind)

React app with React Router in `web/`. API calls proxy through Vite dev server.

**Key structure**:
- `web/src/router/` - Route definitions with `ProtectedRoute` wrapper
- `web/src/hooks/useAuth.tsx` - Auth context provider and `useAuth()` hook
- `web/src/hooks/useOrganization.tsx` - Organization context with `currentOrg`, `switchOrg`
- `web/src/hooks/useMembers.ts` - Member management hook
- `web/src/hooks/useInvitations.ts` - Invitation management hooks
- `web/src/hooks/useSubscription.ts` - Subscription and billing hook
- `web/src/types/` - TypeScript types for organization, subscription, etc.
- `web/src/components/organization/` - OrgSwitcher, MemberList, InviteForm, BillingTab
- `web/src/pages/` - Page components (Dashboard, Login, Profile, Organizations, OrgSettings)

**Auth flow**: `AuthProvider` fetches `/auth/me` on load. Use `useAuth()` for `user`, `isAuthenticated`, `isLoading`.

**Organization flow**: `OrganizationProvider` wraps the app. Use `useOrganization()` for `currentOrg`, `organizations`, `switchOrg`.

### Database Schema

| Database   | Schema Management | Location |
|------------|-------------------|----------|
| PostgreSQL | Goose migrations  | `database/migrations/*.sql` |
| DynamoDB   | Terraform         | `infra/dynamodb.tf` |
| Redis      | N/A (session store) | — |

**PostgreSQL Tables**:
- `users` - User accounts (Google OAuth)
- `organizations` - Multi-tenant organizations
- `organization_members` - User-org membership with roles
- `organization_invitations` - Pending invitations
- `projects` - Projects (soft delete via `deleted_at`)
- `subscriptions` - Stripe subscriptions with tier, status, Stripe IDs
