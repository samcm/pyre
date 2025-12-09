.PHONY: all build build-frontend build-backend dev docker-build docker-run docker-up docker-down generate clean test lint

# Default target
all: build

# Build everything
build: build-frontend build-backend

# Build frontend
build-frontend:
	@echo "Building frontend..."
	cd frontend && pnpm install && pnpm build

# Build backend (requires frontend to be built first)
build-backend:
	@echo "Copying frontend dist to backend..."
	rm -rf backend/frontend/dist
	mkdir -p backend/frontend
	cp -r frontend/dist backend/frontend/dist
	@echo "Building backend..."
	cd backend && go build -o ../pyre ./cmd/server

# Run locally
run: build
	./pyre --config config.yaml

# Dev mode - run frontend and backend in parallel
# Ctrl+C will stop both services
dev:
	@echo "Starting dev stack (backend :8080, frontend :5173)..."
	@trap 'kill 0' INT TERM; \
	(cd backend && go run ./cmd/server --config ../config.yaml) & \
	(cd frontend && pnpm dev) & \
	wait

# Run just the backend
dev-backend:
	cd backend && go run ./cmd/server --config ../config.yaml

# Run just the frontend dev server
dev-frontend:
	cd frontend && pnpm dev

# Docker build
docker-build:
	docker build -t pyre:latest .

# Docker run
docker-run:
	docker run -p 8080:8080 \
		-v $(PWD)/config.yaml:/app/config/config.yaml:ro \
		-v $(PWD)/data:/app/data \
		pyre:latest

# Docker compose up
docker-up:
	docker-compose up -d

# Docker compose down
docker-down:
	docker-compose down

# Generate API types
generate:
	@echo "Generating API types..."
	cd backend && oapi-codegen -config oapi-codegen.yaml internal/api/openapi.yaml
	cd frontend && pnpm generate:api

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f pyre
	rm -rf frontend/dist
	rm -rf frontend/node_modules
	rm -rf backend/frontend
	rm -rf data/*.db

# Run tests
test:
	@echo "Running backend tests..."
	cd backend && go test ./...

# Run linters
lint:
	@echo "Running backend linter..."
	cd backend && golangci-lint run --new-from-rev="origin/main"
	@echo "Running frontend linter..."
	cd frontend && pnpm lint
