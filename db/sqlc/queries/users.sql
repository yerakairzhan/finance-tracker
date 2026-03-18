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
