package handler

import (
	"context"
	"net/http"

	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/middleware"
	"finance-tracker/pkg/models"
	"finance-tracker/pkg/service"

	"github.com/gin-gonic/gin"
)

type analyticsService interface {
	LastMonthSummary(ctx context.Context, userID int64) (*models.AnalyticsSummary, *apperror.Error)
	DailyProfit(ctx context.Context, userID int64, query models.AnalyticsRangeQuery) ([]models.AnalyticsDailyPoint, *apperror.Error)
	LastMonthExpenseByCategory(ctx context.Context, userID int64) ([]models.AnalyticsCategoryExpense, *apperror.Error)
	MonthlyProfit(ctx context.Context, userID int64, query models.AnalyticsMonthlyProfitQuery) ([]models.AnalyticsMonthlyProfitPoint, *apperror.Error)
}

type AnalyticsHandler struct {
	analyticsService analyticsService
}

func NewAnalyticsHandler(analyticsService *service.AnalyticsService) *AnalyticsHandler {
	if analyticsService == nil {
		return &AnalyticsHandler{}
	}
	return &AnalyticsHandler{analyticsService: analyticsService}
}

// LastMonthSummary godoc
// @Summary Last month summary
// @Description Income, expense and profit for the previous calendar month.
// @Tags analytics
// @Security BearerAuth
// @Produce json
// @Success 200 {object} AnalyticsSummary
// @Failure 401 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/analytics/summary/last-month [get]
func (h *AnalyticsHandler) LastMonthSummary(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	out, appErr := h.analyticsService.LastMonthSummary(c.Request.Context(), userID)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, out)
}

// DailyProfit godoc
// @Summary Daily profit series
// @Description Daily income/expense/profit for range (default: previous month).
// @Tags analytics
// @Security BearerAuth
// @Produce json
// @Param from query string false "Start date YYYY-MM-DD"
// @Param to query string false "End date YYYY-MM-DD"
// @Success 200 {array} AnalyticsDailyPoint
// @Failure 400 {object} ErrorEnvelope
// @Failure 401 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/analytics/daily-profit [get]
func (h *AnalyticsHandler) DailyProfit(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}

	var query models.AnalyticsRangeQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		writeError(c, apperror.Validation(err.Error()))
		return
	}
	out, appErr := h.analyticsService.DailyProfit(c.Request.Context(), userID, query)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, out)
}

// LastMonthExpenseByCategory godoc
// @Summary Last month expense categories
// @Description Expense distribution by category for the previous calendar month.
// @Tags analytics
// @Security BearerAuth
// @Produce json
// @Success 200 {array} AnalyticsCategoryExpense
// @Failure 401 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/analytics/expense-categories/last-month [get]
func (h *AnalyticsHandler) LastMonthExpenseByCategory(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	out, appErr := h.analyticsService.LastMonthExpenseByCategory(c.Request.Context(), userID)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, out)
}

// MonthlyProfit godoc
// @Summary Monthly profit trend
// @Description Month-by-month income/expense/profit trend for frontend charts.
// @Tags analytics
// @Security BearerAuth
// @Produce json
// @Param months query int false "Number of months, default 6, max 24" default(6)
// @Success 200 {array} AnalyticsMonthlyProfitPoint
// @Failure 400 {object} ErrorEnvelope
// @Failure 401 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/analytics/monthly-profit [get]
func (h *AnalyticsHandler) MonthlyProfit(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	var query models.AnalyticsMonthlyProfitQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		writeError(c, apperror.Validation(err.Error()))
		return
	}
	out, appErr := h.analyticsService.MonthlyProfit(c.Request.Context(), userID, query)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, out)
}
