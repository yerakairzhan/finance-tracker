package handler

import (
	"net/http"

	"finance-tracker/pkg/models"
	"finance-tracker/pkg/repository"
	"github.com/gin-gonic/gin"
)

// AccountHandler handles account-related HTTP requests
type AccountHandler struct {
	repo *repository.AccountRepository
}

// NewAccountHandler creates a new AccountHandler
func NewAccountHandler(repo *repository.AccountRepository) *AccountHandler {
	return &AccountHandler{repo: repo}
}

// Create creates a new financial account
// POST /accounts
func (h *AccountHandler) Create(c *gin.Context) {
	var req models.CreateAccountRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	account, err := h.repo.CreateAccount(
		c.Request.Context(),
		req.UserID,
		req.AccountType,
		req.Balance,
		req.Currency,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}
	
	c.JSON(http.StatusCreated, account)
}
