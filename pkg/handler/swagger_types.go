package handler

import "finance-tracker/pkg/models"

type ErrorBody struct {
	Code    string `json:"code" example:"VALIDATION_ERROR"`
	Message string `json:"message" example:"invalid account id"`
}

type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

type StatusResponse struct {
	Status string `json:"status" example:"ok"`
}

type (
	AuthTokens               = models.AuthTokens
	RegisterRequest          = models.RegisterRequest
	LoginRequest             = models.LoginRequest
	RefreshRequest           = models.RefreshRequest
	LogoutRequest            = models.LogoutRequest
	UpdateMeRequest          = models.UpdateMeRequest
	ChangePasswordRequest    = models.ChangePasswordRequest
	CreateAccountRequest     = models.CreateAccountRequest
	UpdateAccountRequest     = models.UpdateAccountRequest
	ListTransactionsQuery    = models.ListTransactionsQuery
	AnalyticsRangeQuery      = models.AnalyticsRangeQuery
	AnalyticsMonthlyProfitQuery = models.AnalyticsMonthlyProfitQuery
	CreateTransactionRequest = models.CreateTransactionRequest
	UpdateTransactionRequest = models.UpdateTransactionRequest
	User                     = models.User
	Account                  = models.Account
	Transaction              = models.Transaction
	AnalyticsSummary         = models.AnalyticsSummary
	AnalyticsDailyPoint      = models.AnalyticsDailyPoint
	AnalyticsCategoryExpense = models.AnalyticsCategoryExpense
	AnalyticsMonthlyProfitPoint = models.AnalyticsMonthlyProfitPoint
)
