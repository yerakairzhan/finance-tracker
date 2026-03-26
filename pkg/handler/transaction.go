package handler

import (
	"context"
	"net/http"
	"strconv"

	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/middleware"
	"finance-tracker/pkg/models"
	"finance-tracker/pkg/service"
	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	txService transactionService
}

type transactionService interface {
	List(ctx context.Context, userID int64, query models.ListTransactionsQuery) ([]models.Transaction, *apperror.Error)
	Create(ctx context.Context, userID int64, req models.CreateTransactionRequest) (*models.Transaction, *apperror.Error)
	GetByID(ctx context.Context, userID, txID int64) (*models.Transaction, *apperror.Error)
	Update(ctx context.Context, userID, txID int64, req models.UpdateTransactionRequest) (*models.Transaction, *apperror.Error)
	Delete(ctx context.Context, userID, txID int64) *apperror.Error
}

func NewTransactionHandler(txService *service.TransactionService) *TransactionHandler {
	if txService == nil {
		return &TransactionHandler{}
	}
	return &TransactionHandler{txService: txService}
}

// List godoc
// @Summary List transactions
// @Description List authenticated user's transactions with filters.
// @Tags transactions
// @Security BearerAuth
// @Produce json
// @Param account_id query int false "Account ID"
// @Param category_id query int false "Category ID"
// @Param type query string false "income|expense|transfer"
// @Param from query string false "Start date YYYY-MM-DD"
// @Param to query string false "End date YYYY-MM-DD"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Page size (max 100)" default(20)
// @Success 200 {array} Transaction
// @Failure 400 {object} ErrorEnvelope
// @Failure 401 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/transactions [get]
func (h *TransactionHandler) List(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	var query models.ListTransactionsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		writeError(c, apperror.Validation(err.Error()))
		return
	}
	out, appErr := h.txService.List(c.Request.Context(), userID, query)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Create godoc
// @Summary Create transaction
// @Description Create a transaction for authenticated user.
// @Tags transactions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateTransactionRequest true "Create transaction payload"
// @Success 201 {object} Transaction
// @Failure 400 {object} ErrorEnvelope
// @Failure 401 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/transactions [post]
func (h *TransactionHandler) Create(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	var req models.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, apperror.Validation(err.Error()))
		return
	}
	out, appErr := h.txService.Create(c.Request.Context(), userID, req)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusCreated, out)
}

// GetByID godoc
// @Summary Get transaction
// @Description Get transaction by id for authenticated user.
// @Tags transactions
// @Security BearerAuth
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} Transaction
// @Failure 400 {object} ErrorEnvelope
// @Failure 401 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Router /api/v1/transactions/{id} [get]
func (h *TransactionHandler) GetByID(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(c, apperror.Validation("invalid transaction id"))
		return
	}
	out, appErr := h.txService.GetByID(c.Request.Context(), userID, id)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Update godoc
// @Summary Update transaction
// @Description Update amount/category/notes for authenticated user transaction.
// @Tags transactions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Transaction ID"
// @Param request body UpdateTransactionRequest true "Update transaction payload"
// @Success 200 {object} Transaction
// @Failure 400 {object} ErrorEnvelope
// @Failure 401 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/transactions/{id} [patch]
func (h *TransactionHandler) Update(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(c, apperror.Validation("invalid transaction id"))
		return
	}
	var req models.UpdateTransactionRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		writeError(c, apperror.Validation(err.Error()))
		return
	}
	out, appErr := h.txService.Update(c.Request.Context(), userID, id, req)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Delete godoc
// @Summary Delete transaction
// @Description Soft-delete transaction for authenticated user.
// @Tags transactions
// @Security BearerAuth
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} ErrorEnvelope
// @Failure 401 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/transactions/{id} [delete]
func (h *TransactionHandler) Delete(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(c, apperror.Validation("invalid transaction id"))
		return
	}
	if appErr := h.txService.Delete(c.Request.Context(), userID, id); appErr != nil {
		writeError(c, appErr)
		return
	}
	c.Status(http.StatusNoContent)
}
