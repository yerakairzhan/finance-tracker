package handler

import (
	"net/http"
	"strconv"

	"finance-tracker/pkg/repository"
	"github.com/gin-gonic/gin"
)

// TransactionHandler handles transaction-related HTTP requests
type TransactionHandler struct {
	repo *repository.TransactionRepository
}

// NewTransactionHandler creates a new TransactionHandler
func NewTransactionHandler(repo *repository.TransactionRepository) *TransactionHandler {
	return &TransactionHandler{repo: repo}
}

// List returns a list of transactions for an account
// GET /transactions
func (h *TransactionHandler) List(c *gin.Context) {
	accountIDStr := c.Query("account_id")
	if accountIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account_id is required"})
		return
	}

	accountID, err := strconv.Atoi(accountIDStr)
	if err != nil || accountID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account_id must be a positive integer"})
		return
	}

	// Parse limit and offset with sane defaults.
	limit := 50
	offset := 0

	if l := c.Query("limit"); l != "" {
		parsedLimit, err := strconv.Atoi(l)
		if err != nil || parsedLimit <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be a positive integer"})
			return
		}
		limit = parsedLimit
	}

	if o := c.Query("offset"); o != "" {
		parsedOffset, err := strconv.Atoi(o)
		if err != nil || parsedOffset < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "offset must be a non-negative integer"})
			return
		}
		offset = parsedOffset
	}

	transactions, err := h.repo.ListTransactionsByAccountID(c.Request.Context(), accountID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}
	
	c.JSON(http.StatusOK, transactions)
}
