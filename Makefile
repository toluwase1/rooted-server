.PHONY: dev db migrate run build test clean

# Start local databases (PostgreSQL + Redis)
db:
	docker compose up -d

# Stop local databases
db-stop:
	docker compose down

# Run database migrations
migrate:
	psql "postgres://postgres:postgres@localhost:5432/rooted?sslmode=disable" -f migrations/001_initial.up.sql

# Rollback migrations
migrate-down:
	psql "postgres://postgres:postgres@localhost:5432/rooted?sslmode=disable" -f migrations/001_initial.down.sql

# Run the server in development
run:
	go run ./cmd/server

# Build the binary
build:
	go build -o bin/rooted-server ./cmd/server

# Run tests
test:
	go test ./... -v

# Clean build artifacts
clean:
	rm -rf bin/

# Full local dev setup: start DBs, run migrations, start server
dev: db
	@echo "Waiting for databases to start..."
	@sleep 3
	@make migrate
	@make run
