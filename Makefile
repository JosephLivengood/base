.PHONY: dev dev-down dev-logs dev-rebuild build clean migrate migrate-down migrate-status migrate-new tf-init tf-plan tf-apply tf-plan-prod tf-apply-prod

# Start everything for development
dev:
	docker compose up -d --build

# Stop all services
dev-down:
	docker compose down

# View logs
dev-logs:
	docker compose logs -f

# Rebuild and restart
dev-rebuild:
	docker compose up -d --build --force-recreate

# Build API binary locally
build:
	cd api && go build -o bin/server .

# Clean build artifacts and docker volumes
clean:
	rm -rf api/bin
	docker compose down -v

# Database URL for migrations
DB_URL ?= postgres://postgres:postgres@localhost:5432/base?sslmode=disable

# Run pending migrations
migrate:
	docker compose run --rm goose up

# Rollback last migration
migrate-down:
	docker compose run --rm goose down

# Show migration status
migrate-status:
	docker compose run --rm goose status

# Create new migration (usage: make migrate-new name=create_users_table)
migrate-new:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-new name=migration_name"; exit 1; fi
	docker compose run --rm goose create $(name) sql

# Terraform - Initialize
tf-init:
	cd infra && terraform init

# Terraform - Plan (local)
tf-plan:
	cd infra && terraform plan -var-file=environments/local.tfvars

# Terraform - Apply (local)
tf-apply:
	cd infra && terraform apply -var-file=environments/local.tfvars

# Terraform - Plan (production)
tf-plan-prod:
	cd infra && terraform plan -var-file=environments/prod.tfvars

# Terraform - Apply (production)
tf-apply-prod:
	cd infra && terraform apply -var-file=environments/prod.tfvars
