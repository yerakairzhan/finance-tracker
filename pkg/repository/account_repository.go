package repository

import (
	"context"

	"finance-tracker/db/queries"
	"finance-tracker/pkg/models"
)

type AccountRepository struct {
	q *queries.Queries
}

func NewAccountRepository(q *queries.Queries) *AccountRepository {
	return &AccountRepository{q: q}
}


/*
Helper: convert sqlc row -> models.Account
*/
func mapAccount(row sqlc.Account) models.Account {
	return models.Account{
		ID:          int(row.ID),
		UserID:      int(row.UserID),
		AccountType: row.AccountType,
		Balance:     numericToString(row.Balance),
		Currency:    row.Currency,
		CreatedAt:   timestampToTime(row.CreatedAt),
		UpdatedAt:   timestampToTime(row.UpdatedAt),
	}
}

func (ar *AccountRepository) Create(
	ctx context.Context,
	userID int,
	accType string,
	currency string,
	balance string,
) (*models.Account, error) {

	num, err := stringToNumeric(balance)
	if err != nil {
		return nil, err
	}
	
	row, err := ar.q.CreateAccount(ctx, sqlc.CreateAccountParams{
		UserID:      int32(userID),
		AccountType: accType,
		Balance:     num,
		Currency:    currency,
	})

	if err != nil {
		return nil, err
	}

	account := mapAccount(row)
	return &account, nil
}

func (ar *AccountRepository) GetByID(ctx context.Context, id int) (*models.Account, error) {

	row, err := ar.q.GetAccountByID(ctx, int32(id))
	if err != nil {
		return nil, err
	}

	account := mapAccount(row)
	return &account, nil
}

func (ar *AccountRepository) List(ctx context.Context) ([]models.Account, error) {

	rows, err := ar.q.ListAccounts(ctx)
	if err != nil {
		return nil, err
	}

	accounts := make([]models.Account, 0, len(rows))

	for _, row := range rows {
		accounts = append(accounts, mapAccount(row))
	}

	return accounts, nil
}

func (ar *AccountRepository) GetByUserID(ctx context.Context, userID int) ([]models.Account, error) {

	rows, err := ar.q.GetAccountsByUserID(ctx, int32(userID))
	if err != nil {
		return nil, err
	}

	accounts := make([]models.Account, 0, len(rows))

	for _, row := range rows {
		accounts = append(accounts, mapAccount(row))
	}

	return accounts, nil
}

func (ar *AccountRepository) Delete(ctx context.Context, id int) error {
	return ar.q.DeleteAccount(ctx, int32(id))
}

func (ar *AccountRepository) GetBalance(ctx context.Context, id int) (string, error) {

	num, err := ar.q.GetAccountBalance(ctx, int32(id))
	if err != nil {
		return "", err
	}

	return numericToString(num), nil
}