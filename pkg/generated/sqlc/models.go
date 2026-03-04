package sqlc

import (
	"database/sql"
	"time"
)

type User struct {
	ID        int32
	Email     string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Account struct {
	ID          int32
	UserID      int32
	AccountType string
	Balance     string
	Currency    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Transaction struct {
	ID              int32
	AccountID       int32
	Amount          string
	Description     sql.NullString
	TransactionType string
	CreatedAt       time.Time
}
