# Finance Tracker API (Go Monolith)

Backend service for personal finance tracking with authentication, account/transaction management, and chart-ready analytics endpoints.

## Stack

- Go + Gin
- PostgreSQL
- Redis
- sqlc + pgx
- Docker / Docker Compose
- Swagger (OpenAPI)

## Main Features

- JWT auth with register/login/refresh/logout
- Refresh token is transported only via `HttpOnly` cookie (never in JSON)
- Access-token revocation in Redis on logout
- Redis-backed refresh session rotation (one-time use semantics)
- User profile endpoints (`me`, update profile, change password)
- Accounts CRUD with soft-delete
- Transactions CRUD with ownership checks and account balance recalculation
- Analytics endpoints for frontend charts:
  - last-month summary (income/expense/profit)
  - daily profit series
  - last-month expense categories
  - monthly profit trend (N months)

## API Surface

Base prefix: `/api/v1`

- Auth
  - `POST /api/v1/auth/register`
  - `POST /api/v1/auth/login` (sets refresh cookie)
  - `POST /api/v1/auth/refresh` (reads refresh cookie, rotates it, sets new one)
  - `POST /api/v1/auth/logout` (JWT, clears refresh cookie)
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
- Analytics
  - `GET /api/v1/analytics/summary/last-month` (JWT)
  - `GET /api/v1/analytics/daily-profit` (JWT, optional `from`/`to`)
  - `GET /api/v1/analytics/expense-categories/last-month` (JWT)
  - `GET /api/v1/analytics/monthly-profit` (JWT, query `months=1..24`, default `6`)
- Health
  - `GET /health`
  - `GET /health/ready`

Swagger UI: `http://localhost:8080/docs/index.html`

## Business Rules

- Monetary values are `numeric(15,4)` in DB and returned as strings.
- Accounts and transactions use soft-delete (`deleted_at`).
- JWT access token TTL is 15 minutes.
- Refresh token TTL is 30 days and is stored in Redis as **hash only**.
- Logout revokes the current access token in Redis until token expiry.
- User ownership is enforced in DB-layer queries.
- Refresh token is never returned in JSON and never accepted via JSON body.
- Auth responses are sanitized server-side:
  - `refresh_token` and internal session/token metadata are removed.
- Refresh rotation is atomic; old refresh token cannot be reused (replay-safe).

Error format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "..."
  }
}
```

## Run with Docker

```bash
make docker-run
```

Services:
- API: `http://localhost:8080`
- Postgres: `localhost:5435`
- Redis: `localhost:6379`

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

## Frontend Submodule (important)

Frontend is tracked as a Git submodule (`frontend`), not a regular folder.

Initialize/sync it from backend root:

```bash
git submodule sync --recursive
git submodule update --init --recursive
```

If local `frontend` is broken/diverged:

```bash
git submodule deinit -f frontend
rm -rf frontend .git/modules/frontend
git submodule update --init --recursive frontend
```

Update frontend to latest `main` (optional; this changes submodule pointer in backend repo):

```bash
cd frontend
git fetch origin
git checkout main
git pull --ff-only origin main
cd ..
git add frontend
git commit -m "chore: bump frontend submodule"
```

Run all services (including frontend if defined in compose):

```bash
docker compose up -d --build
# or
make docker-run
```

## Quick API Examples

### Auth flow (cookie-based refresh)

```bash
# Login: stores cookies to cookie jar
curl -i -c cookies.txt -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"password123"}'

# Refresh: sends cookies from cookie jar (no JSON body required)
curl -i -b cookies.txt -c cookies.txt -X POST http://localhost:8080/api/v1/auth/refresh

# Logout: JWT + cookies; clears refresh cookie server-side
curl -i -b cookies.txt -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

### Accounts

```bash
curl -s -X POST http://localhost:8080/api/v1/accounts \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Kaspi","account_type":"bank_card","currency":"KZT","balance":"100000.0000"}'
```

### Transactions

```bash
curl -s -X POST http://localhost:8080/api/v1/transactions \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"account_id":1,"amount":"12000.0000","currency":"KZT","type":"expense","description":"Groceries","transacted_at":"2026-03-01"}'
```

### Redis revocation example

```bash
# Logout revokes access token in Redis and deletes refresh session
curl -i -b cookies.txt -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

## Analytics Examples (for frontend)

### 1) Last month summary (profit for previous month)

```bash
curl -s http://localhost:8080/api/v1/analytics/summary/last-month \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

Example response:

```json
{
  "period_start": "2026-02-01",
  "period_end": "2026-02-28",
  "income": "125000.0000",
  "expense": "84320.5000",
  "profit": "40679.5000"
}
```

### 2) Daily chart series

```bash
curl -s "http://localhost:8080/api/v1/analytics/daily-profit?from=2026-02-01&to=2026-02-28" \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

Example response:

```json
[
  {"date":"2026-02-01","income":"1000.0000","expense":"200.0000","profit":"800.0000"},
  {"date":"2026-02-02","income":"0.0000","expense":"350.0000","profit":"-350.0000"}
]
```

### 3) Expense pie/donut by category

```bash
curl -s http://localhost:8080/api/v1/analytics/expense-categories/last-month \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

### 4) Monthly trend chart

```bash
curl -s "http://localhost:8080/api/v1/analytics/monthly-profit?months=12" \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

Example response:

```json
[
  {"month":"2025-04","income":"100000.0000","expense":"70000.0000","profit":"30000.0000"},
  {"month":"2025-05","income":"120000.0000","expense":"85000.0000","profit":"35000.0000"}
]
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
pkg/cache/               # redis integration
pkg/models/              # request/response models
```

## Docs

- API endpoints: `docs/api-endpoints.md`
- DB schema: `docs/db-schema.md`
