# Finance Tracker API (Go Monolith)

Backend service for personal finance tracking.  
Current implemented scope is **v1 core**: auth, users, accounts, transactions, and health checks.

## Stack

- Go + Gin
- PostgreSQL
- Redis
- sqlc + pgx
- Docker / Docker Compose
- Swagger (OpenAPI)

## Current API Surface

Base prefix: `/api/v1`

- Auth
  - `POST /api/v1/auth/register`
  - `POST /api/v1/auth/login`
  - `POST /api/v1/auth/refresh`
  - `POST /api/v1/auth/logout` (JWT)
- Users
  - `GET /api/v1/users/me` (JWT)
  - `PATCH /api/v1/users/me` (JWT)
  - `PATCH /api/v1/users/me/password` (JWT)
- Accounts
  - `GET /api/v1/accounts` (JWT)
  - `POST /api/v1/accounts` (JWT)
  - `GET /api/v1/accounts/:id` (JWT)
  - `PATCH /api/v1/accounts/:id` (JWT)
  - `DELETE /api/v1/accounts/:id` (JWT, soft-delete)
- Transactions
  - `GET /api/v1/transactions` (JWT)
  - `POST /api/v1/transactions` (JWT)
  - `GET /api/v1/transactions/:id` (JWT)
  - `PATCH /api/v1/transactions/:id` (JWT)
  - `DELETE /api/v1/transactions/:id` (JWT, soft-delete)
- Health
  - `GET /health`
  - `GET /health/ready`

Swagger UI: `http://localhost:8080/docs/index.html`

## Key Rules Implemented

- Monetary values are `numeric(15,4)` in DB and returned as strings (no float64).
- Accounts and transactions use soft-delete (`deleted_at`).
- JWT access token TTL: 15 minutes.
- Refresh token TTL: 30 days; stored as bcrypt hash in `refresh_tokens`.
- Logout revokes access token in Redis until JWT expiry.
- User ownership is enforced using JWT `user_id` in DB queries.
- Unified error envelope:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "..."
  }
}
```

## Project Structure

```text
cmd/api/                 # entrypoint
internal/app/            # app bootstrapping + route wiring
db/migrations/           # SQL migrations
db/sqlc/schema/          # schema used by sqlc
db/sqlc/queries/         # query definitions for sqlc
db/queries/              # generated sqlc code
pkg/handler/             # HTTP handlers
pkg/service/             # business logic
pkg/repository/          # DB access wrappers
pkg/middleware/          # auth middleware
pkg/models/              # request/response models
```

## Run with Docker

```bash
make docker-run
```

API: `http://localhost:8080`  
DB: `localhost:5435`

The API now applies the embedded SQL migrations automatically on startup, so a fresh `docker compose up --build` should come up without any manual DB prep.

Stop:

```bash
make docker-stop
```

## Local Development

```bash
export DATABASE_URL=postgres://postgres:postgres@localhost:5435/finance_tracker?sslmode=disable
export PORT=8080
export JWT_SECRET=change-me
export REDIS_ADDR=localhost:6379
export REDIS_PASSWORD=

go mod tidy
make sqlc
go run ./cmd/api
```

## Redis Integration Examples

### 1) Start dependencies (Postgres + Redis + API)

```bash
make docker-run
```

### 2) Local run with explicit Redis config

```bash
REDIS_ADDR=localhost:6379 REDIS_PASSWORD="" DATABASE_URL=postgres://postgres:postgres@localhost:5435/finance_tracker?sslmode=disable PORT=8080 JWT_SECRET=change-me go run ./cmd/api
```

### 3) Behavior example: logout invalidates current access token

```bash
# login (example payload)
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"password123"}'

# call protected endpoint with access token -> 200
curl -i http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <ACCESS_TOKEN>"

# logout with refresh token + same bearer access token -> 204
curl -i -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -d '{"refresh_token":"<REFRESH_TOKEN>"}'

# same access token is now revoked in Redis -> 401
curl -i http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

## Docs

- API endpoints: `docs/api-endpoints.md`
- DB schema: `docs/db-schema.md`
