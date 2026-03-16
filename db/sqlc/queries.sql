-- name: CreateUser :one
INSERT INTO users (email, name)
VALUES ($1, $2)
RETURNING id, email, name, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, email, name, created_at, updated_at
FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT id, email, name, created_at, updated_at
FROM users
ORDER BY id;

-- name: UpdateUser :one
UPDATE users
SET email = $2, name = $3, updated_at = NOW()
WHERE id = $1
RETURNING id, email, name, created_at, updated_at;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: CreateAccount :one
INSERT INTO accounts (user_id, account_type, balance, currency)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, account_type, balance, currency, created_at, updated_at;

-- name: GetAccountsByUserID :many
SELECT id, user_id, account_type, balance, currency, created_at, updated_at
FROM accounts
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CreateTransaction :one
INSERT INTO transactions (account_id, amount, description, transaction_type)
VALUES ($1, $2, $3, $4)
RETURNING id, account_id, amount, description, transaction_type, created_at;

-- name: ListTransactionsByAccountID :many
SELECT id, account_id, amount, description, transaction_type, created_at
FROM transactions
WHERE account_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
