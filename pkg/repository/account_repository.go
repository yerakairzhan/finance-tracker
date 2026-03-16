package repository

import (
	"context"

	"finance-tracker/pkg/generated/sqlc"
	"finance-tracker/pkg/models"
)

type AccountRepository struct {
	q *sqlc.Queries
}

func NewAccountRepository(q *sqlc.Queries) *AccountRepository {
	return &AccountRepository{q: q}
}

func (ar *AccountRepository) CreateAccount(
	ctx context.Context,
	userID int,
	accountType,
	balance,
	currency string,
) (*models.Account, error) {

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
