.PHONY: dev db migrate run build test test-matching test-user lint check deploy clean

# ==========================================
# LOCAL DEV
# ==========================================

db:
	docker compose up -d

db-stop:
	docker compose down

migrate:
	psql "postgres://postgres:postgres@localhost:5432/rooted?sslmode=disable" -f migrations/001_initial.up.sql
	psql "postgres://postgres:postgres@localhost:5432/rooted?sslmode=disable" -f migrations/002_admin_auth.up.sql

run:
	go run ./cmd/server

dev: db
	@echo "Waiting for databases..."
	@sleep 3
	@make migrate
	@make run

# ==========================================
# BUILD & TEST
# ==========================================

build:
	go build -o bin/rooted-server ./cmd/server

test:
	go test ./internal/... -timeout 180s -count=1

test-v:
	go test ./internal/... -v -timeout 180s -count=1

test-matching:
	go test ./internal/matching/ -v -timeout 120s -count=1

test-user:
	go test ./internal/user/ -v -timeout 120s -count=1

lint:
	go vet ./...

# Build + lint + test — run before every push
check: lint build test
	@echo "✅ All checks passed"

# ==========================================
# FRONTEND
# ==========================================

build-miniapp:
	cd miniapp && npx tsc --noEmit && VITE_API_URL=https://rooted-api-643943133167.us-central1.run.app npx vite build

build-admin:
	cd admin && npx tsc --noEmit && VITE_API_URL=https://rooted-api-643943133167.us-central1.run.app npx vite build

# ==========================================
# GIT (SSH key is toluwasethomas, repo is toluwase1 — use gh auth via HTTPS)
# ==========================================

push:
	GIT_CONFIG_GLOBAL=/dev/null git -c credential.helper='!gh auth git-credential' push https://github.com/toluwase1/rooted-server.git main

pull:
	GIT_CONFIG_GLOBAL=/dev/null git -c credential.helper='!gh auth git-credential' pull https://github.com/toluwase1/rooted-server.git main

# ==========================================
# DEPLOY
# ==========================================

deploy-api:
	source .env.deploy && ./scripts/deploy-api.sh

deploy-userbot:
	./scripts/deploy-userbot.sh

deploy-miniapp: build-miniapp
	npx wrangler pages deploy miniapp/dist --project-name=rooted-miniapp --commit-dirty=true

deploy-admin: build-admin
	npx wrangler pages deploy admin/dist --project-name=rooted-admin --commit-dirty=true

deploy-all: check deploy-api deploy-userbot deploy-miniapp deploy-admin
	@echo "✅ All deployed"

# ==========================================
# CLEANUP
# ==========================================

clean:
	rm -rf bin/ miniapp/dist/ admin/dist/
