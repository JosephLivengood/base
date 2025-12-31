.PHONY: dev dev-down dev-logs dev-rebuild build clean

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
