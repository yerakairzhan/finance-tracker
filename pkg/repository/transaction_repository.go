package repository

import (
	"context"

	"finance-tracker/db/queries"
	"finance-tracker/pkg/models"
)

type TransactionRepository struct {
	q *queries.Queries
}

func NewTransactionRepository(q *queries.Queries) *TransactionRepository {
	return &TransactionRepository{q: q}
}

func (tr *TransactionRepository) CreateTransaction(
	ctx context.Context,
	accountID int,
	amount,
	description,
	txType string,
) (*models.Transaction, error) {

	num, err := stringToNumeric(amount)
	if err != nil {
		return nil, err
	}

	row, err := tr.q.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		AccountID:       int32(accountID),
		Amount:          num,
		Description:     description,
		TransactionType: txType,
	})

	if err != nil {
		return nil, err
	}

	return &models.Transaction{
		ID:              int(row.ID),
		AccountID:       int(row.AccountID),
		Amount:          numericToString(row.Amount),
		Description:     row.Description,
		TransactionType: row.TransactionType,
		CreatedAt:       timestampToTime(row.CreatedAt),
	}, nil
}

// ListTransactionsByAccountID retrieves transactions for an account
func (tr *TransactionRepository) ListTransactionsByAccountID(ctx context.Context, accountID, limit, offset int) ([]models.Transaction, error) {
	rows, err := tr.q.ListTransactionsByAccountID(ctx, queries.ListTransactionsByAccountIDParams{
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
			Amount:          numericToString(row.Amount),
			Description:     row.Description,
			TransactionType: row.TransactionType,
			CreatedAt:       timestampToTime(row.CreatedAt),
		})
	}
	return transactions, nil
}

func (tr *TransactionRepository) GetByID(ctx context.Context, id int) (*models.Transaction, error) {
	row, err := tr.q.GetTransactionByID(ctx, int32(id))
	if err != nil {
		return nil, err
	}

	return &models.Transaction{
		ID:              int(row.ID),
		AccountID:       int(row.AccountID),
		Amount:          numericToString(row.Amount),
		Description:     row.Description,
		TransactionType: row.TransactionType,
		CreatedAt:       timestampToTime(row.CreatedAt),
	}, nil
}

func (tr *TransactionRepository) List(ctx context.Context, limit, offset int) ([]models.Transaction, error) {
	rows, err := tr.q.ListTransactions(ctx, sqlc.ListTransactionsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	var transactions []models.Transaction
	for _, row := range rows {
		transactions = append(transactions, models.Transaction{
			ID:              int(row.ID),
			AccountID:       int(row.AccountID),
			Amount:          numericToString(row.Amount),
			Description:     row.Description,
			TransactionType: row.TransactionType,
			CreatedAt:       timestampToTime(row.CreatedAt),
		})
	}

	return transactions, nil
}

func (tr *TransactionRepository) Delete(ctx context.Context, id int) error {
	return tr.q.DeleteTransaction(ctx, int32(id))
}

func (tr *TransactionRepository) Search(ctx context.Context, query string, limit, offset int) ([]models.Transaction, error) {
	rows, err := tr.q.SearchTransactions(ctx, sqlc.SearchTransactionsParams{
		Description: stringToText(query),
		Limit:       int32(limit),
		Offset:      int32(offset),
	})
	if err != nil {
		return nil, err
	}

	var transactions []models.Transaction
	for _, row := range rows {
		transactions = append(transactions, models.Transaction{
			ID:              int(row.ID),
			AccountID:       int(row.AccountID),
			Amount:          numericToString(row.Amount),
			Description:     row.Description,
			TransactionType: row.TransactionType,
			CreatedAt:       timestampToTime(row.CreatedAt),
		})
	}

	return transactions, nil
}
