package repository

import (
	"context"

	"finance-tracker/pkg/generated/sqlc"
	"finance-tracker/pkg/models"
)

type TransactionRepository struct {
	q *sqlc.Queries
}

func NewTransactionRepository(q *sqlc.Queries) *TransactionRepository {
	return &TransactionRepository{q: q}
}

func (tr *TransactionRepository) CreateTransaction(
	ctx context.Context,
	accountID int,
	amount,
	description,
	txType string,
) (*models.Transaction, error) {

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
		Description:     row.Description,
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
			Description:     row.Description,
			TransactionType: row.TransactionType,
			CreatedAt:       row.CreatedAt,
		})
	}
	return transactions, nil
}