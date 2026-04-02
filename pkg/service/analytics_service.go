package service

import (
	"context"
	"fmt"
	"time"

	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/models"
	"finance-tracker/pkg/repository"
)

type AnalyticsService struct {
	txRepo *repository.TransactionRepository
}

func NewAnalyticsService(txRepo *repository.TransactionRepository) *AnalyticsService {
	return &AnalyticsService{txRepo: txRepo}
}

func (s *AnalyticsService) LastMonthSummary(ctx context.Context, userID int64) (*models.AnalyticsSummary, *apperror.Error) {
	start, end := lastMonthRangeUTC(time.Now().UTC())
	row, err := s.txRepo.LastMonthSummary(ctx, userID, start, end)
	if err != nil {
		return nil, apperror.Internal("failed to load analytics summary")
	}
	return &models.AnalyticsSummary{
		PeriodStart: start.Format("2006-01-02"),
		PeriodEnd:   end.Format("2006-01-02"),
		Income:      numericToString4(row.Income),
		Expense:     numericToString4(row.Expense),
		Profit:      numericToString4(row.Profit),
	}, nil
}

func (s *AnalyticsService) DailyProfit(ctx context.Context, userID int64, query models.AnalyticsRangeQuery) ([]models.AnalyticsDailyPoint, *apperror.Error) {
	start, end, err := rangeFromQuery(query)
	if err != nil {
		return nil, apperror.Validation(err.Error())
	}
	rows, err := s.txRepo.DailyProfit(ctx, userID, start, end)
	if err != nil {
		return nil, apperror.Internal("failed to load analytics daily series")
	}
	out := make([]models.AnalyticsDailyPoint, 0, len(rows))
	for _, row := range rows {
		out = append(out, models.AnalyticsDailyPoint{
			Date:    row.Date.Format("2006-01-02"),
			Income:  numericToString4(row.Income),
			Expense: numericToString4(row.Expense),
			Profit:  numericToString4(row.Profit),
		})
	}
	return out, nil
}

func (s *AnalyticsService) LastMonthExpenseByCategory(ctx context.Context, userID int64) ([]models.AnalyticsCategoryExpense, *apperror.Error) {
	start, end := lastMonthRangeUTC(time.Now().UTC())
	rows, err := s.txRepo.LastMonthExpenseByCategory(ctx, userID, start, end)
	if err != nil {
		return nil, apperror.Internal("failed to load analytics categories")
	}
	out := make([]models.AnalyticsCategoryExpense, 0, len(rows))
	for _, row := range rows {
		out = append(out, models.AnalyticsCategoryExpense{
			Category: row.Category,
			Amount:   numericToString4(row.Amount),
		})
	}
	return out, nil
}

func (s *AnalyticsService) MonthlyProfit(ctx context.Context, userID int64, query models.AnalyticsMonthlyProfitQuery) ([]models.AnalyticsMonthlyProfitPoint, *apperror.Error) {
	months := query.Months
	if months <= 0 {
		months = 6
	}
	now := time.Now().UTC()
	endMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	startMonth := endMonth.AddDate(0, -(months - 1), 0)

	rows, err := s.txRepo.MonthlyProfit(ctx, userID, startMonth, endMonth)
	if err != nil {
		return nil, apperror.Internal("failed to load monthly profit")
	}

	out := make([]models.AnalyticsMonthlyProfitPoint, 0, len(rows))
	for _, row := range rows {
		out = append(out, models.AnalyticsMonthlyProfitPoint{
			Month:   row.Month.Format("2006-01"),
			Income:  numericToString4(row.Income),
			Expense: numericToString4(row.Expense),
			Profit:  numericToString4(row.Profit),
		})
	}
	return out, nil
}

func rangeFromQuery(query models.AnalyticsRangeQuery) (time.Time, time.Time, error) {
	if query.From == nil && query.To == nil {
		start, end := lastMonthRangeUTC(time.Now().UTC())
		return start, end, nil
	}

	if query.From == nil || query.To == nil {
		return time.Time{}, time.Time{}, fmt.Errorf("both from and to are required when one is provided")
	}

	start, err := time.Parse("2006-01-02", *query.From)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid from date, expected YYYY-MM-DD")
	}
	end, err := time.Parse("2006-01-02", *query.To)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid to date, expected YYYY-MM-DD")
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("to must be greater than or equal to from")
	}
	if end.Sub(start) > 366*24*time.Hour {
		return time.Time{}, time.Time{}, fmt.Errorf("date range is too large (max 366 days)")
	}
	return start.UTC(), end.UTC(), nil
}

func lastMonthRangeUTC(now time.Time) (time.Time, time.Time) {
	firstOfCurrentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	firstOfLastMonth := firstOfCurrentMonth.AddDate(0, -1, 0)
	lastOfLastMonth := firstOfCurrentMonth.AddDate(0, 0, -1)
	return firstOfLastMonth, lastOfLastMonth
}
