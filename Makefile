# Check if the environment file exists
ENVFILE := .env
ifneq ("$(wildcard $(ENVFILE))","")
	include $(ENVFILE)
	export $(shell sed 's/=.*//' $(ENVFILE))
endif

# Gobot Makefile
EXECUTABLE=gobot

.PHONY: help dev build build-cli run clean test deps gen setup sqlc migrate-status migrate-up migrate-down cli

# Default target
help:
	@echo "Gobot - Full-stack SaaS Boilerplate"
	@echo ""
	@echo "Quick Start:"
	@echo "  make dev       - Start everything (backend + frontend)"
	@echo ""
	@echo "Development:"
	@echo "  make air       - Backend only with hot reload"
	@echo "  make build     - Build production binary"
	@echo "  make test      - Run tests"
	@echo "  make gen       - Regenerate API code from .api file"
	@echo ""
	@echo "CLI Agent:"
	@echo "  make build-cli - Build the CLI agent"
	@echo "  make cli       - Build and install CLI globally"
	@echo ""
	@echo "Database:"
	@echo "  make migrate-up     - Run pending migrations"
	@echo "  make migrate-down   - Rollback last migration"
	@echo "  make migrate-status - Check migration status"
	@echo ""
	@echo "Other:"
	@echo "  make deps      - Download dependencies"
	@echo "  make clean     - Clean build artifacts"
	@echo "  make setup NAME=myapp - Rename project"

# Start everything - the easiest way to develop
dev:
	@echo "Starting Gobot development environment..."
	@echo "Backend: http://localhost:27895"
	@echo "Frontend: http://localhost:27458"
	@echo ""
	@if command -v docker >/dev/null 2>&1 && (docker compose version >/dev/null 2>&1 || docker-compose version >/dev/null 2>&1); then \
		echo "Using Docker..."; \
		docker compose up; \
	else \
		echo "Starting without Docker (use two terminals for better experience)..."; \
		echo ""; \
		$(MAKE) air & \
		sleep 2; \
		cd app && pnpm dev; \
	fi

# Build the application
build:
	@echo "Building $(EXECUTABLE)..."
	go build -o bin/$(EXECUTABLE) .

# Build the CLI agent
build-cli:
	@echo "Building gobot CLI agent..."
	go build -o bin/gobot-cli ./cmd/gobot-cli

# Install CLI globally
cli: build-cli
	@echo "Installing gobot-cli..."
	cp bin/gobot-cli $(GOPATH)/bin/gobot 2>/dev/null || cp bin/gobot-cli /usr/local/bin/gobot 2>/dev/null || echo "Copy bin/gobot-cli to your PATH manually"
	@echo "Done! Run 'gobot --help' to get started"

# Run the application
run: build
	@echo "Starting $(EXECUTABLE)..."
	./bin/$(EXECUTABLE)

# Run with air (hot reload)
air:
	@echo "Starting with hot reload..."
	air

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf tmp/

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Code generation from .api file
gen:
	@echo "Cleaning auto-generated handlers..."
	@rm -rf internal/handler
	@echo "Generating Go API code..."
	goctl api go -api $(EXECUTABLE).api -dir . --style gozero
	@echo "Generating TypeScript API code..."
	goctl api ts -api $(EXECUTABLE).api -dir ./app/src/lib/api/
	@echo "Code generation complete!"

# Setup script - rename project
setup:
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make setup NAME=myapp"; \
		exit 1; \
	fi
	@echo "Renaming project from gobot to $(NAME)..."
	@./install.sh $(NAME)
	@echo "Project renamed to $(NAME)!"
	@echo "Run 'make deps && make gen' to complete setup"

# Database migrations and code generation
sqlc:
	@echo "Generating sqlc code..."
	sqlc generate
	@echo "sqlc code generation complete!"

migrate-status:
	@echo "Checking migration status..."
	go run cmd/migrate/main.go status

migrate-up:
	@echo "Running migrations..."
	go run cmd/migrate/main.go up

migrate-down:
	@echo "Rolling back last migration..."
	go run cmd/migrate/main.go down

# Docker commands
docker-build:
	@echo "Building Docker image..."
	docker build -t $(EXECUTABLE) .

docker-run:
	@echo "Running Docker container..."
	docker run -p 27895:27895 --env-file .env $(EXECUTABLE)

# Development environment
dev-setup: deps
	@echo "Setting up development environment..."
	@mkdir -p bin
	@cd app && pnpm install
	@echo "Development setup complete!"
	@echo "Run 'make gen' to generate API code"
	@echo "Run 'make run' to start the backend"
	@echo "Run 'cd app && pnpm dev' to start the frontend"
