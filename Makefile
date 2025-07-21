.PHONY: help build start stop restart logs clean test dev setup quick-start db-migrate db-seed

# Default target
help:
	@echo "CloudBox Development Commands"
	@echo "============================="
	@echo "build     - Build all Docker images"
	@echo "start     - Start all services"
	@echo "stop      - Stop all services"
	@echo "restart   - Restart all services"
	@echo "logs      - Show logs for all services"
	@echo "clean     - Clean up containers and volumes"
	@echo "test      - Run tests"
	@echo "dev       - Start development environment"

# Build Docker images
build:
	docker-compose build

# Start all services
start:
	docker-compose up -d

# Stop all services
stop:
	docker-compose down

# Restart all services
restart: stop start

# Show logs
logs:
	docker-compose logs -f

# Clean up everything
clean:
	docker-compose down -v
	docker system prune -f

# Run tests
test:
	cd backend && go test ./...

# Development mode (with logs)
dev:
	docker-compose up

# Setup development environment
setup:
	@echo "Setting up CloudBox development environment..."
	@cp .env.example .env
	@echo "Created .env file from .env.example"
	@echo "Please edit .env file with your configuration"
	@echo "Then run: make build && make start"

# Database operations
db-migrate:
	docker-compose exec backend go run cmd/migrate/main.go

db-seed:
	docker-compose exec backend go run cmd/seed/main.go

# Quick start for new developers
quick-start: setup build start
	@echo "CloudBox is now running!"
	@echo "Frontend: http://localhost:3000"
	@echo "Backend API: http://localhost:8080"
	@echo "Health check: http://localhost:8080/health"