
-- name: CreateTransaction :one
INSERT INTO transactions (account_id, amount, description, transaction_type)
VALUES ($1, $2, $3, $4)
RETURNING id, account_id, amount, description, transaction_type, created_at;

-- name: CreateTransactionWithBalanceUpdate :one
WITH new_tx AS (
    INSERT INTO transactions (account_id, amount, description, transaction_type)
    VALUES ($1, $2, $3, $4)
    RETURNING *
)
UPDATE accounts
SET balance = balance + 
    CASE 
        WHEN $4 = 'income' THEN $2
        ELSE -$2
    END
WHERE account_id = $1
RETURNING (SELECT * FROM new_tx);

-- name: ListTransactionsByAccountID :many
SELECT id, account_id, amount, description, transaction_type, created_at
FROM transactions
WHERE account_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListTransactions :many
SELECT *
FROM transactions
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetTransactionByID :one
SELECT *
FROM transactions
WHERE id = $1;

-- name: SearchTransactions :many
SELECT *
FROM transactions
WHERE description ILIKE '%' || sqlc.arg('Description') || '%'
ORDER BY created_at DESC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Offset');

-- name: DeleteTransaction :exec
DELETE FROM transactions
WHERE id = $1;

-- name: ExportTransactions :many
SELECT *
FROM transactions
ORDER BY created_at DESC;

-- name: FilterTransactionsByType :many
SELECT *
FROM transactions
WHERE transaction_type = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: FilterTransactionsByDate :many
SELECT *
FROM transactions
WHERE created_at BETWEEN $1 AND $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: AdvancedSearchTransactions :many
SELECT *
FROM transactions
WHERE ($1::text IS NULL OR description ILIKE '%' || $1 || '%')
  AND ($2::int IS NULL OR account_id = $2)
  AND ($3::text IS NULL OR transaction_type = $3)
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;