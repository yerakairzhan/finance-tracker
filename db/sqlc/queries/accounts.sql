-- name: CreateAccount :one
INSERT INTO accounts (user_id, account_type, balance, currency)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, account_type, balance, currency, created_at, updated_at;

-- name: GetAccountByID :one
SELECT id, user_id, account_type, balance, currency, created_at, updated_at
FROM accounts
WHERE id = $1;

-- name: ListAccounts :many
SELECT id, user_id, account_type, balance, currency, created_at, updated_at
FROM accounts
ORDER BY created_at DESC;

-- name: GetAccountsByUserID :many
SELECT id, user_id, account_type, balance, currency, created_at, updated_at
FROM accounts
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateAccount :one
UPDATE accounts
SET account_type = $2,
    currency = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING id, user_id, account_type, balance, currency, created_at, updated_at;

-- name: DeleteAccount :exec
DELETE FROM accounts
WHERE id = $1;

-- name: GetAccountBalance :one
SELECT balance
FROM accounts
WHERE id = $1;