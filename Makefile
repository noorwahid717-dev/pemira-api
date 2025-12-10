.PHONY: help dev build test lint docker-up docker-down db-restore db-verify

help:
	@echo "Available commands:"
	@echo "  make dev             - Run development server"
	@echo "  make build           - Build the application"
	@echo "  make test            - Run tests"
	@echo "  make lint            - Run linter"
	@echo "  make docker-up       - Start docker services"
	@echo "  make docker-down     - Stop docker services"
	@echo "  make db-restore      - Restore database from backup"
	@echo "  make db-verify       - Verify database connection and data"

dev:
	@echo "Starting development server..."
	@go run cmd/api/main.go

build:
	@echo "Building application..."
	@go build -o build/api cmd/api/main.go

test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...

lint:
	@echo "Running linter..."
	@golangci-lint run

docker-up:
	@echo "Starting docker services..."
	@docker-compose up -d

docker-down:
	@echo "Stopping docker services..."
	@docker-compose down

db-restore:
	@echo "Restoring database from backup..."
	@./restore_db.sh

db-verify:
	@echo "Verifying database connection..."
	@PGPASSWORD="AZcIF926bLLeeVRQ" psql -h aws-1-ap-southeast-1.pooler.supabase.com -p 6543 -U postgres.xqzfrodnznhjstfstvyz -d postgres -c "SELECT 'elections' AS table_name, COUNT(*) AS rows FROM myschema.elections UNION ALL SELECT 'voters', COUNT(*) FROM myschema.voters UNION ALL SELECT 'candidates', COUNT(*) FROM myschema.candidates UNION ALL SELECT 'votes', COUNT(*) FROM myschema.votes;"
