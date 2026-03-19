.PHONY: help run dev test tidy fmt sqlc swagger gen docker-up docker-down docker-logs docker-run docker-stop migrate

APP_PKG := ./cmd/api
SWAG    := go run github.com/swaggo/swag/cmd/swag@v1.16.6

# Defaults (override: `make run PORT=9090 DATABASE_URL=...`)
PORT ?= 8080
DATABASE_URL ?= postgres://postgres:postgres@localhost:5435/finance_tracker?sslmode=disable

help:
	@printf "%s\n" "Targets:"
	@printf "%s\n" "  run          Run API locally (uses DATABASE_URL, PORT)"
	@printf "%s\n" "  dev          Run with air (requires air installed)"
	@printf "%s\n" "  test         Run tests"
	@printf "%s\n" "  tidy         go mod tidy"
	@printf "%s\n" "  fmt          gofmt -w on repo"
	@printf "%s\n" "  sqlc         sqlc generate"
	@printf "%s\n" "  swagger      regenerate docs/ (swag init)"
	@printf "%s\n" "  gen          sqlc + swagger"
	@printf "%s\n" "  docker-up    docker compose up --build"
	@printf "%s\n" "  docker-down  docker compose down"
	@printf "%s\n" "  docker-logs  docker compose logs -f"
	@printf "%s\n" "  docker-run   docker compose up --build -d"
	@printf "%s\n" "  docker-stop  docker compose down"
	@printf "%s\n" "  migrate      apply db/migrations/001_init.sql using psql and DATABASE_URL"

run:
	@DATABASE_URL="$(DATABASE_URL)" PORT="$(PORT)" go run $(APP_PKG)

dev:
	@command -v air >/dev/null 2>&1 || { echo "air not found. Install: go install github.com/air-verse/air@latest"; exit 1; }
	@DATABASE_URL="$(DATABASE_URL)" PORT="$(PORT)" air

test:
	@go test ./...

tidy:
	@go mod tidy

fmt:
	@gofmt -w .

sqlc:
	@command -v sqlc >/dev/null 2>&1 || { echo "sqlc not found. Install: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"; exit 1; }
	@sqlc generate

swagger:
	@$(SWAG) init -g internal/app/app.go -o docs --parseInternal --parseDependency

gen: sqlc swagger

docker-up:
	@docker compose up --build

docker-down:
	@docker compose down

docker-logs:
	@docker compose logs -f

docker-run:
	@docker compose up --build -d

docker-stop: docker-down

migrate:
	@command -v psql >/dev/null 2>&1 || { echo "psql not found. Install Postgres client tools."; exit 1; }
	@psql "$(DATABASE_URL)" -f db/migrations/001_init.sql
