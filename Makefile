# Makefile for Order Processing System

.PHONY: build run tidy docker-up docker-down migrate seed clean help

# Build the application
build:
	go build -o bin/server ./cmd/src/server

# Run the application locally
run:
	go run cmd/src/server/main.go

# Resolve dependencies
tidy:
	go mod tidy

# Start RabbitMQ via Docker
docker-up:
	docker-compose up -d

# Stop RabbitMQ
docker-down:
	docker-compose down

# Run the seeding script
seed:
	go run seed_products.go

# Clean build artifacts
clean:
	if exist bin rmdir /s /q bin
	if exist orders.db del orders.db

# Help command
help:
	@echo "Available commands:"
	@echo "  make build       - Build the server binary"
	@echo "  make run         - Run the server locally"
	@echo "  make tidy        - Tidy go modules"
	@echo "  make docker-up   - Start RabbitMQ infrastructure"
	@echo "  make docker-down - Stop infrastructure"
	@echo "  make seed        - Seed initial products into database"
	@echo "  make clean       - Remove build artifacts"
