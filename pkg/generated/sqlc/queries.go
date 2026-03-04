package sqlc

import (
	"context"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (email, name)
VALUES ($1, $2)
RETURNING id, email, name, created_at, updated_at
`

func (q *Queries) CreateUser(ctx context.Context, email string, name string) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser, email, name)
	var i User
	err := row.Scan(&i.ID, &i.Email, &i.Name, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

const getUserByID = `-- name: GetUserByID :one
SELECT id, email, name, created_at, updated_at
FROM users
WHERE id = $1
`

func (q *Queries) GetUserByID(ctx context.Context, id int32) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByID, id)
	var i User
	err := row.Scan(&i.ID, &i.Email, &i.Name, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

type CreateAccountParams struct {
	UserID      int32
	AccountType string
	Balance     string
	Currency    string
}

const createAccount = `-- name: CreateAccount :one
INSERT INTO accounts (user_id, account_type, balance, currency)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, account_type, balance, currency, created_at, updated_at
`

func (q *Queries) CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error) {
	row := q.db.QueryRowContext(ctx, createAccount, arg.UserID, arg.AccountType, arg.Balance, arg.Currency)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.AccountType,
		&i.Balance,
		&i.Currency,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getAccountsByUserID = `-- name: GetAccountsByUserID :many
SELECT id, user_id, account_type, balance, currency, created_at, updated_at
FROM accounts
WHERE user_id = $1
ORDER BY created_at DESC
`

func (q *Queries) GetAccountsByUserID(ctx context.Context, userID int32) ([]Account, error) {
	rows, err := q.db.QueryContext(ctx, getAccountsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Account
	for rows.Next() {
		var i Account
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.AccountType,
			&i.Balance,
			&i.Currency,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

type CreateTransactionParams struct {
	AccountID       int32
	Amount          string
	Description     string
	TransactionType string
}

const createTransaction = `-- name: CreateTransaction :one
INSERT INTO transactions (account_id, amount, description, transaction_type)
VALUES ($1, $2, $3, $4)
RETURNING id, account_id, amount, description, transaction_type, created_at
`

func (q *Queries) CreateTransaction(ctx context.Context, arg CreateTransactionParams) (Transaction, error) {
	row := q.db.QueryRowContext(ctx, createTransaction, arg.AccountID, arg.Amount, arg.Description, arg.TransactionType)
	var i Transaction
	err := row.Scan(
		&i.ID,
		&i.AccountID,
		&i.Amount,
		&i.Description,
		&i.TransactionType,
		&i.CreatedAt,
	)
	return i, err
}

type ListTransactionsByAccountIDParams struct {
	AccountID int32
	Limit     int32
	Offset    int32
}

const listTransactionsByAccountID = `-- name: ListTransactionsByAccountID :many
SELECT id, account_id, amount, description, transaction_type, created_at
FROM transactions
WHERE account_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3
`

func (q *Queries) ListTransactionsByAccountID(ctx context.Context, arg ListTransactionsByAccountIDParams) ([]Transaction, error) {
	rows, err := q.db.QueryContext(ctx, listTransactionsByAccountID, arg.AccountID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Transaction
	for rows.Next() {
		var i Transaction
		if err := rows.Scan(
			&i.ID,
			&i.AccountID,
			&i.Amount,
			&i.Description,
			&i.TransactionType,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
