package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"finance-tracker/pkg/repository"
)

type AccountHandler struct {
	repo *repository.AccountRepository
}

func NewAccountHandler(r *repository.AccountRepository) *AccountHandler {
	return &AccountHandler{repo: r}
}

// Create godoc
// @Summary Create account
// @Description Create a new financial account
// @Tags accounts
// @Accept json
// @Produce json
// @Param request body CreateAccountRequest true "Create account payload"
// @Success 201 {object} Account
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /accounts [post]
func (h *AccountHandler) Create(c *gin.Context) {

	var req struct {
		UserID      int    `json:"user_id"`
		AccountType string `json:"account_type"`
		Currency    string `json:"currency"`
		Balance     string `json:"balance"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	account, err := h.repo.Create(
		c.Request.Context(),
		req.UserID,
		req.AccountType,
		req.Currency,
		req.Balance,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, account)
}

func (h *AccountHandler) GetByID(c *gin.Context) {

	id, _ := strconv.Atoi(c.Param("id"))

	account, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, account)
}

func (h *AccountHandler) List(c *gin.Context) {

	accounts, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, accounts)
}

func (h *AccountHandler) GetUserAccounts(c *gin.Context) {

	id, _ := strconv.Atoi(c.Param("id"))

	accounts, err := h.repo.GetByUserID(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, accounts)
}

func (h *AccountHandler) Delete(c *gin.Context) {

	id, _ := strconv.Atoi(c.Param("id"))

	err := h.repo.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "account deleted"})
}

func (h *AccountHandler) GetBalance(c *gin.Context) {

	id, _ := strconv.Atoi(c.Param("id"))

	balance, err := h.repo.GetBalance(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"balance": balance})
}