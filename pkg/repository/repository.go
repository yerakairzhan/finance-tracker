package repository

import (
	"context"

	"github.com/example/financial-intelligence-platform/pkg/generated/sqlc"
	"github.com/example/financial-intelligence-platform/pkg/models"
)

// UserRepository handles user-related database operations
type UserRepository struct {
	q *sqlc.Queries
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(q *sqlc.Queries) *UserRepository {
	return &UserRepository{q: q}
}

// CreateUser inserts a new user into the database
func (ur *UserRepository) CreateUser(ctx context.Context, email, name string) (*models.User, error) {
	row, err := ur.q.CreateUser(ctx, email, name)
	if err != nil {
		return nil, err
	}
	return &models.User{
		ID:        int(row.ID),
		Email:     row.Email,
		Name:      row.Name,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

// GetUserByID retrieves a user by ID
func (ur *UserRepository) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	row, err := ur.q.GetUserByID(ctx, int32(id))
	if err != nil {
		return nil, err
	}
	return &models.User{
		ID:        int(row.ID),
		Email:     row.Email,
		Name:      row.Name,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

// AccountRepository handles account-related database operations
type AccountRepository struct {
	q *sqlc.Queries
}

// NewAccountRepository creates a new AccountRepository instance
func NewAccountRepository(q *sqlc.Queries) *AccountRepository {
	return &AccountRepository{q: q}
}

// CreateAccount inserts a new account into the database
func (ar *AccountRepository) CreateAccount(ctx context.Context, userID int, accountType, balance, currency string) (*models.Account, error) {
	row, err := ar.q.CreateAccount(ctx, sqlc.CreateAccountParams{
		UserID:      int32(userID),
		AccountType: accountType,
		Balance:     balance,
		Currency:    currency,
	})
	if err != nil {
		return nil, err
	}
	return &models.Account{
		ID:          int(row.ID),
		UserID:      int(row.UserID),
		AccountType: row.AccountType,
		Balance:     row.Balance,
		Currency:    row.Currency,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

// GetAccountsByUserID retrieves all accounts for a user
func (ar *AccountRepository) GetAccountsByUserID(ctx context.Context, userID int) ([]models.Account, error) {
	rows, err := ar.q.GetAccountsByUserID(ctx, int32(userID))
	if err != nil {
		return nil, err
	}
	var accounts []models.Account
	for _, row := range rows {
		accounts = append(accounts, models.Account{
			ID:          int(row.ID),
			UserID:      int(row.UserID),
			AccountType: row.AccountType,
			Balance:     row.Balance,
			Currency:    row.Currency,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		})
	}
	return accounts, nil
}

// TransactionRepository handles transaction-related database operations
type TransactionRepository struct {
	q *sqlc.Queries
}

// NewTransactionRepository creates a new TransactionRepository instance
func NewTransactionRepository(q *sqlc.Queries) *TransactionRepository {
	return &TransactionRepository{q: q}
}

// CreateTransaction inserts a new transaction into the database
func (tr *TransactionRepository) CreateTransaction(ctx context.Context, accountID int, amount, description, txType string) (*models.Transaction, error) {
	row, err := tr.q.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		AccountID:       int32(accountID),
		Amount:          amount,
		Description:     description,
		TransactionType: txType,
	})
	if err != nil {
		return nil, err
	}
	return &models.Transaction{
		ID:              int(row.ID),
		AccountID:       int(row.AccountID),
		Amount:          row.Amount,
		Description:     row.Description.String,
		TransactionType: row.TransactionType,
		CreatedAt:       row.CreatedAt,
	}, nil
}

// ListTransactionsByAccountID retrieves transactions for an account
func (tr *TransactionRepository) ListTransactionsByAccountID(ctx context.Context, accountID, limit, offset int) ([]models.Transaction, error) {
	rows, err := tr.q.ListTransactionsByAccountID(ctx, sqlc.ListTransactionsByAccountIDParams{
		AccountID: int32(accountID),
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}
	var transactions []models.Transaction
	for _, row := range rows {
		transactions = append(transactions, models.Transaction{
			ID:              int(row.ID),
			AccountID:       int(row.AccountID),
			Amount:          row.Amount,
			Description:     row.Description.String,
			TransactionType: row.TransactionType,
			CreatedAt:       row.CreatedAt,
		})
	}
	return transactions, nil
}
