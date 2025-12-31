.PHONY: help proto migrate-up migrate-down run test docker-up docker-down clean build install-tools mongo-init

help:
	@echo "Available commands:"
	@echo "  make install-tools  - Install required tools (buf, migrate, etc.)"
	@echo "  make proto          - Generate code from proto files"
	@echo "  make migrate-up     - Run database migrations up"
	@echo "  make migrate-down   - Run database migrations down"
	@echo "  make mongo-init     - Initialize MongoDB collections"
	@echo "  make run            - Run the application"
	@echo "  make build          - Build the application"
	@echo "  make test           - Run tests"
	@echo "  make docker-up      - Start all services with Docker Compose"
	@echo "  make docker-down    - Stop all services"
	@echo "  make clean          - Clean build artifacts"

install-tools:
	@echo "Installing required tools..."
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Tools installed successfully"

proto:
	@echo "Generating code from proto files..."
	buf generate
	@echo "Proto generation complete"

migrate-up:
	@echo "Running migrations up..."
	migrate -path migrations/postgres -database "postgresql://support_user:support_pass@localhost:5432/support_db?sslmode=disable" up
	@echo "Migrations complete"

migrate-down:
	@echo "Running migrations down..."
	migrate -path migrations/postgres -database "postgresql://support_user:support_pass@localhost:5432/support_db?sslmode=disable" down
	@echo "Migrations rolled back"

mongo-init:
	@echo "Initializing MongoDB..."
	mongosh support_db < migrations/mongodb/init.js
	@echo "MongoDB initialized"

run:
	@echo "Starting application..."
	go run cmd/server/main.go

build:
	@echo "Building application..."
	go build -o bin/server cmd/server/main.go
	@echo "Build complete: bin/server"

test:
	@echo "Running tests..."
	go test -v ./...

docker-up:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d
	@echo "Services started"
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo "Running migrations..."
	@make migrate-up
	@echo "Initializing MongoDB..."
	@docker-compose exec -T mongodb mongosh support_db < migrations/mongodb/init.js
	@echo "Setup complete"

docker-down:
	@echo "Stopping services..."
	docker-compose down
	@echo "Services stopped"

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf gen/
	@echo "Clean complete"

.env:
	@if [ ! -f .env ]; then \
		echo "Creating .env file from .env.example..."; \
		cp .env.example .env; \
	fi
