# Base

Application starter template with multi-tenant organizations, projects, and Stripe subscriptions.

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
│   │   │   ├── organization/ # Multi-tenant orgs, members, invitations
│   │   │   ├── project/    # Projects (belong to organizations)
│   │   │   ├── subscription/ # Stripe subscriptions
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
│       ├── components/     # Auth, layout, organization components
│       ├── hooks/          # useAuth, useOrganization, useSubscription
│       ├── types/          # TypeScript types
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
| PostgreSQL | Users, orgs, projects, subscriptions | Goose migrations (`database/migrations/`) |
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

## Organizations

Multi-tenant organization system with role-based access control.

**Features:**
- Users automatically get a "Personal" organization on signup
- Roles: `owner`, `admin`, `member`
- Invitation system with email-based invites
- Organization switcher in UI

**Endpoints:**
```
GET    /api/organizations              # List user's organizations
POST   /api/organizations              # Create organization
GET    /api/organizations/:id          # Get organization
PUT    /api/organizations/:id          # Update organization
DELETE /api/organizations/:id          # Delete organization (owner only)
POST   /api/organizations/:id/leave    # Leave organization
POST   /api/organizations/active       # Set active organization

# Members
GET    /api/organizations/:id/members          # List members
PUT    /api/organizations/:id/members/:userId  # Update member role
DELETE /api/organizations/:id/members/:userId  # Remove member

# Invitations
GET    /api/organizations/:id/invitations      # List org invitations
POST   /api/organizations/:id/invitations      # Create invitation
DELETE /api/organizations/:id/invitations/:id  # Cancel invitation
GET    /api/invitations                        # User's pending invitations
POST   /api/invitations/:id/accept             # Accept invitation
POST   /api/invitations/:id/decline            # Decline invitation
```

## Projects

Projects belong to organizations and support soft delete.

**Endpoints:**
```
GET    /api/organizations/:orgId/projects      # List projects
POST   /api/organizations/:orgId/projects      # Create project
GET    /api/organizations/:orgId/projects/:id  # Get project
PUT    /api/organizations/:orgId/projects/:id  # Update project
DELETE /api/organizations/:orgId/projects/:id  # Soft delete project
```

## Subscriptions & Billing

Stripe-powered subscription billing with two tiers: Free and Personal Plus.

**Features:**
- Stripe Checkout for payment collection
- Stripe Billing Portal for subscription management
- Optimistic updates in development (no webhooks needed locally)
- Webhook-driven updates in production

**Tiers:**
| Tier | Price |
|------|-------|
| Free | $0 |
| Personal Plus | $10/month or $95/year |

**Endpoints:**
```
GET  /api/organizations/:orgId/subscription          # Get subscription
POST /api/organizations/:orgId/subscription/checkout # Create checkout session
POST /api/organizations/:orgId/subscription/portal   # Create billing portal session
POST /api/organizations/:orgId/subscription/confirm  # Confirm checkout (dev)
POST /api/stripe/webhook                             # Stripe webhook (no auth)
```

**Environment Variables:**
```bash
STRIPE_SECRET_KEY=sk_test_...
STRIPE_PUBLISHABLE_KEY=pk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...  # Optional for local dev
```

**Local Development:**
For local testing, checkout confirmation is handled optimistically via the `/confirm` endpoint. In production, configure Stripe webhooks to point to `/api/stripe/webhook`.

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
