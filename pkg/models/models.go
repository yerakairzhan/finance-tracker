package models

import "time"

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Account struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Name        string    `json:"name"`
	AccountType string    `json:"account_type"`
	Balance     string    `json:"balance"`
	Currency    string    `json:"currency"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Transaction struct {
	ID           int64     `json:"id"`
	AccountID    int64     `json:"account_id"`
	CategoryID   *int64    `json:"category_id,omitempty"`
	Amount       string    `json:"amount"`
	Currency     string    `json:"currency"`
	Type         string    `json:"type"`
	Description  string    `json:"description"`
	Notes        *string   `json:"notes,omitempty"`
	TransactedAt string    `json:"transacted_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=128"`
	Name     string `json:"name" binding:"required,min=1,max=120"`
	Currency string `json:"currency" binding:"required,len=3,uppercase"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required,min=32"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required,min=32"`
}

type UpdateMeRequest struct {
	Name     *string `json:"name" binding:"omitempty,min=1,max=120"`
	Currency *string `json:"currency" binding:"omitempty,len=3,uppercase"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,min=8,max=128"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=128"`
}

type CreateAccountRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=120"`
	AccountType string `json:"account_type" binding:"required,oneof=cash bank_card e_wallet"`
	Currency    string `json:"currency" binding:"required,len=3,uppercase"`
	Balance     string `json:"balance" binding:"required"`
}

type UpdateAccountRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=120"`
	AccountType *string `json:"account_type" binding:"omitempty,oneof=cash bank_card e_wallet"`
	Currency    *string `json:"currency" binding:"omitempty,len=3,uppercase"`
}

type ListTransactionsQuery struct {
	AccountID  *int64  `form:"account_id" binding:"omitempty,min=1"`
	CategoryID *int64  `form:"category_id" binding:"omitempty,min=1"`
	Type       *string `form:"type" binding:"omitempty,oneof=income expense transfer"`
	From       *string `form:"from" binding:"omitempty,datetime=2006-01-02"`
	To         *string `form:"to" binding:"omitempty,datetime=2006-01-02"`
	Page       int     `form:"page,default=1" binding:"min=1"`
	Limit      int     `form:"limit,default=20" binding:"min=1,max=100"`
}

type CreateTransactionRequest struct {
	AccountID    int64   `json:"account_id" binding:"required,min=1"`
	CategoryID   *int64  `json:"category_id" binding:"omitempty,min=1"`
	Amount       string  `json:"amount" binding:"required"`
	Currency     string  `json:"currency" binding:"required,len=3,uppercase"`
	Type         string  `json:"type" binding:"required,oneof=income expense transfer"`
	Description  string  `json:"description" binding:"required,min=1,max=255"`
	Notes        *string `json:"notes" binding:"omitempty,max=1000"`
	TransactedAt string  `json:"transacted_at" binding:"required,datetime=2006-01-02"`
}

type UpdateTransactionRequest struct {
	Amount     *string `json:"amount" binding:"omitempty"`
	CategoryID *int64  `json:"category_id" binding:"omitempty,min=1"`
	Notes      *string `json:"notes" binding:"omitempty,max=1000"`
}

type AnalyticsRangeQuery struct {
	From *string `form:"from" binding:"omitempty,datetime=2006-01-02"`
	To   *string `form:"to" binding:"omitempty,datetime=2006-01-02"`
}

type AnalyticsSummary struct {
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
	Income      string `json:"income"`
	Expense     string `json:"expense"`
	Profit      string `json:"profit"`
}

type AnalyticsDailyPoint struct {
	Date    string `json:"date"`
	Income  string `json:"income"`
	Expense string `json:"expense"`
	Profit  string `json:"profit"`
}

type AnalyticsCategoryExpense struct {
	Category string `json:"category"`
	Amount   string `json:"amount"`
}

type AnalyticsMonthlyProfitQuery struct {
	Months int `form:"months,default=6" binding:"min=1,max=24"`
}

type AnalyticsMonthlyProfitPoint struct {
	Month  string `json:"month"`
	Income string `json:"income"`
	Expense string `json:"expense"`
	Profit string `json:"profit"`
}
