package models

import "time"

// User represents a user in the system
type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Account represents a financial account
type Account struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	AccountType   string    `json:"account_type"`
	Balance       string    `json:"balance"` // Using string for decimal
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Transaction represents a financial transaction
type Transaction struct {
	ID              int       `json:"id"`
	AccountID       int       `json:"account_id"`
	Amount          string    `json:"amount"` // Using string for decimal
	Description     string    `json:"description"`
	TransactionType string    `json:"transaction_type"`
	CreatedAt       time.Time `json:"created_at"`
}

// Request/Response DTOs

// RegisterRequest is the request body for user registration
type RegisterRequest struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name" binding:"required"`
}

// CreateAccountRequest is the request body for creating an account
type CreateAccountRequest struct {
	UserID      int    `json:"user_id" binding:"required"`
	AccountType string `json:"account_type" binding:"required"`
	Balance     string `json:"balance" binding:"required"`
	Currency    string `json:"currency" binding:"required"`
}

type UpdateUserRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// ListTransactionsRequest is the query for listing transactions
type ListTransactionsRequest struct {
	AccountID int `form:"account_id" binding:"required"`
	Limit     int `form:"limit,default=50"`
	Offset    int `form:"offset,default=0"`
}
