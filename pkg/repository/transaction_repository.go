package repository

import (
	"context"
	"fmt"
	"math/big"
	"time"

	sqlc "finance-tracker/db/queries"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

type AnalyticsSummaryRow struct {
	Income  pgtype.Numeric
	Expense pgtype.Numeric
	Profit  pgtype.Numeric
}

type AnalyticsDailyProfitRow struct {
	Date    time.Time
	Income  pgtype.Numeric
	Expense pgtype.Numeric
	Profit  pgtype.Numeric
}

type AnalyticsCategoryExpenseRow struct {
	Category string
	Amount   pgtype.Numeric
}

type AnalyticsMonthlyProfitRow struct {
	Month  time.Time
	Income pgtype.Numeric
	Expense pgtype.Numeric
	Profit pgtype.Numeric
}

func NewTransactionRepository(pool *pgxpool.Pool, q *sqlc.Queries) *TransactionRepository {
	return &TransactionRepository{pool: pool, q: q}
}

func (r *TransactionRepository) ListForUser(ctx context.Context, params sqlc.ListTransactionsForUserParams) ([]sqlc.Transaction, error) {
	return r.q.ListTransactionsForUser(ctx, params)
}

func (r *TransactionRepository) GetByIDForUser(ctx context.Context, txID, userID int64) (sqlc.Transaction, error) {
	return r.q.GetTransactionByIDForUser(ctx, sqlc.GetTransactionByIDForUserParams{
		ID:     txID,
		UserID: userID,
	})
}

func (r *TransactionRepository) CreateForUser(
	ctx context.Context,
	userID int64,
	params sqlc.CreateTransactionParams,
) (sqlc.Transaction, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return sqlc.Transaction{}, err
	}
	defer tx.Rollback(ctx)

	qtx := r.q.WithTx(tx)

	_, err = qtx.GetAccountByIDForUserForUpdate(ctx, sqlc.GetAccountByIDForUserForUpdateParams{
		ID:     params.AccountID,
		UserID: userID,
	})
	if err != nil {
		return sqlc.Transaction{}, err
	}
	if params.CategoryID.Valid {
		if _, err = qtx.CategoryAccessibleForUser(ctx, sqlc.CategoryAccessibleForUserParams{
			ID:     params.CategoryID.Int64,
			UserID: pgtype.Int8{Int64: userID, Valid: true},
		}); err != nil {
			return sqlc.Transaction{}, err
		}
	}

	created, err := qtx.CreateTransaction(ctx, params)
	if err != nil {
		return sqlc.Transaction{}, err
	}

	delta, err := signedAmount(params.Amount, params.Type)
	if err != nil {
		return sqlc.Transaction{}, err
	}
	deltaNum, err := stringToNumeric(delta)
	if err != nil {
		return sqlc.Transaction{}, err
	}

	_, err = qtx.UpdateAccountBalanceDeltaByIDForUser(ctx, sqlc.UpdateAccountBalanceDeltaByIDForUserParams{
		ID:      params.AccountID,
		UserID:  userID,
		Balance: deltaNum,
	})
	if err != nil {
		return sqlc.Transaction{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.Transaction{}, err
	}
	return created, nil
}

func (r *TransactionRepository) UpdateForUser(
	ctx context.Context,
	userID, txID int64,
	params sqlc.UpdateTransactionByIDForUserParams,
) (sqlc.Transaction, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return sqlc.Transaction{}, err
	}
	defer tx.Rollback(ctx)

	qtx := r.q.WithTx(tx)

	current, err := qtx.GetTransactionByIDForUserForUpdate(ctx, sqlc.GetTransactionByIDForUserForUpdateParams{
		ID:     txID,
		UserID: userID,
	})
	if err != nil {
		return sqlc.Transaction{}, err
	}
	if params.CategoryID.Valid {
		if _, err = qtx.CategoryAccessibleForUser(ctx, sqlc.CategoryAccessibleForUserParams{
			ID:     params.CategoryID.Int64,
			UserID: pgtype.Int8{Int64: userID, Valid: true},
		}); err != nil {
			return sqlc.Transaction{}, err
		}
	}

	params.ID = txID
	params.UserID = userID
	updated, err := qtx.UpdateTransactionByIDForUser(ctx, params)
	if err != nil {
		return sqlc.Transaction{}, err
	}

	oldSigned, err := signedAmount(current.Amount, current.Type)
	if err != nil {
		return sqlc.Transaction{}, err
	}
	newSigned, err := signedAmount(updated.Amount, updated.Type)
	if err != nil {
		return sqlc.Transaction{}, err
	}
	delta, err := subtractDecimalStrings(newSigned, oldSigned)
	if err != nil {
		return sqlc.Transaction{}, err
	}

	if delta != "0.0000" {
		deltaNum, err := stringToNumeric(delta)
		if err != nil {
			return sqlc.Transaction{}, err
		}
		_, err = qtx.UpdateAccountBalanceDeltaByIDForUser(ctx, sqlc.UpdateAccountBalanceDeltaByIDForUserParams{
			ID:      current.AccountID,
			UserID:  userID,
			Balance: deltaNum,
		})
		if err != nil {
			return sqlc.Transaction{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.Transaction{}, err
	}
	return updated, nil
}

func (r *TransactionRepository) SoftDeleteForUser(ctx context.Context, userID, txID int64) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := r.q.WithTx(tx)
	current, err := qtx.GetTransactionByIDForUserForUpdate(ctx, sqlc.GetTransactionByIDForUserForUpdateParams{
		ID:     txID,
		UserID: userID,
	})
	if err != nil {
		return err
	}

	affected, err := qtx.SoftDeleteTransactionByIDForUser(ctx, sqlc.SoftDeleteTransactionByIDForUserParams{
		ID:     txID,
		UserID: userID,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return pgx.ErrNoRows
	}

	currentSigned, err := signedAmount(current.Amount, current.Type)
	if err != nil {
		return err
	}
	delta, err := subtractDecimalStrings("0.0000", currentSigned)
	if err != nil {
		return err
	}
	deltaNum, err := stringToNumeric(delta)
	if err != nil {
		return err
	}

	_, err = qtx.UpdateAccountBalanceDeltaByIDForUser(ctx, sqlc.UpdateAccountBalanceDeltaByIDForUserParams{
		ID:      current.AccountID,
		UserID:  userID,
		Balance: deltaNum,
	})
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *TransactionRepository) LastMonthSummary(ctx context.Context, userID int64, start, end time.Time) (AnalyticsSummaryRow, error) {
	const q = `
SELECT
  COALESCE(SUM(CASE WHEN t.type = 'income' THEN t.amount ELSE 0 END), 0) AS income,
  COALESCE(SUM(CASE WHEN t.type = 'expense' THEN t.amount ELSE 0 END), 0) AS expense,
  COALESCE(SUM(CASE WHEN t.type = 'income' THEN t.amount WHEN t.type = 'expense' THEN -t.amount ELSE 0 END), 0) AS profit
FROM transactions t
JOIN accounts a ON a.id = t.account_id
WHERE a.user_id = $1
  AND a.deleted_at IS NULL
  AND t.deleted_at IS NULL
  AND t.transacted_at >= $2::date
  AND t.transacted_at <= $3::date;
`
	var row AnalyticsSummaryRow
	err := r.pool.QueryRow(ctx, q, userID, start.Format("2006-01-02"), end.Format("2006-01-02")).Scan(
		&row.Income,
		&row.Expense,
		&row.Profit,
	)
	return row, err
}

func (r *TransactionRepository) DailyProfit(ctx context.Context, userID int64, start, end time.Time) ([]AnalyticsDailyProfitRow, error) {
	const q = `
SELECT
  d::date AS day,
  COALESCE(SUM(CASE WHEN t.type = 'income' THEN t.amount ELSE 0 END), 0) AS income,
  COALESCE(SUM(CASE WHEN t.type = 'expense' THEN t.amount ELSE 0 END), 0) AS expense,
  COALESCE(SUM(CASE WHEN t.type = 'income' THEN t.amount WHEN t.type = 'expense' THEN -t.amount ELSE 0 END), 0) AS profit
FROM generate_series($2::date, $3::date, interval '1 day') d
LEFT JOIN accounts a
  ON a.user_id = $1
  AND a.deleted_at IS NULL
LEFT JOIN transactions t
  ON t.account_id = a.id
  AND t.transacted_at = d::date
  AND t.deleted_at IS NULL
GROUP BY d
ORDER BY d;
`
	rows, err := r.pool.Query(ctx, q, userID, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]AnalyticsDailyProfitRow, 0)
	for rows.Next() {
		var item AnalyticsDailyProfitRow
		if err = rows.Scan(&item.Date, &item.Income, &item.Expense, &item.Profit); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *TransactionRepository) LastMonthExpenseByCategory(ctx context.Context, userID int64, start, end time.Time) ([]AnalyticsCategoryExpenseRow, error) {
	const q = `
SELECT
  COALESCE(c.name, 'Uncategorized') AS category_name,
  COALESCE(SUM(t.amount), 0) AS amount
FROM transactions t
JOIN accounts a ON a.id = t.account_id
LEFT JOIN categories c ON c.id = t.category_id
WHERE a.user_id = $1
  AND a.deleted_at IS NULL
  AND t.deleted_at IS NULL
  AND t.type = 'expense'
  AND t.transacted_at >= $2::date
  AND t.transacted_at <= $3::date
GROUP BY category_name
ORDER BY amount DESC;
`
	rows, err := r.pool.Query(ctx, q, userID, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]AnalyticsCategoryExpenseRow, 0)
	for rows.Next() {
		var item AnalyticsCategoryExpenseRow
		if err = rows.Scan(&item.Category, &item.Amount); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *TransactionRepository) MonthlyProfit(ctx context.Context, userID int64, startMonth, endMonth time.Time) ([]AnalyticsMonthlyProfitRow, error) {
	const q = `
SELECT
  m.month_start::date AS month_start,
  COALESCE(SUM(CASE WHEN t.type = 'income' THEN t.amount ELSE 0 END), 0) AS income,
  COALESCE(SUM(CASE WHEN t.type = 'expense' THEN t.amount ELSE 0 END), 0) AS expense,
  COALESCE(SUM(CASE WHEN t.type = 'income' THEN t.amount WHEN t.type = 'expense' THEN -t.amount ELSE 0 END), 0) AS profit
FROM generate_series($2::date, $3::date, interval '1 month') AS m(month_start)
LEFT JOIN accounts a
  ON a.user_id = $1
  AND a.deleted_at IS NULL
LEFT JOIN transactions t
  ON t.account_id = a.id
  AND t.deleted_at IS NULL
  AND date_trunc('month', t.transacted_at)::date = m.month_start::date
GROUP BY m.month_start
ORDER BY m.month_start;
`
	rows, err := r.pool.Query(ctx, q, userID, startMonth.Format("2006-01-02"), endMonth.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]AnalyticsMonthlyProfitRow, 0)
	for rows.Next() {
		var item AnalyticsMonthlyProfitRow
		if err = rows.Scan(&item.Month, &item.Income, &item.Expense, &item.Profit); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func signedAmount(amount pgtype.Numeric, txType string) (string, error) {
	raw := numericToString4(amount)
	if txType == "income" {
		return raw, nil
	}
	if txType == "expense" || txType == "transfer" {
		if raw == "0.0000" {
			return raw, nil
		}
		return "-" + raw, nil
	}
	return "", fmt.Errorf("invalid transaction type")
}

func subtractDecimalStrings(left, right string) (string, error) {
	l, ok := new(big.Rat).SetString(left)
	if !ok {
		return "", fmt.Errorf("invalid decimal value")
	}
	r, ok := new(big.Rat).SetString(right)
	if !ok {
		return "", fmt.Errorf("invalid decimal value")
	}
	return new(big.Rat).Sub(l, r).FloatString(4), nil
}
