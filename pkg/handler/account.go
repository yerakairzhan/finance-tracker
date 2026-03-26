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

type AccountHandler struct {
	accountService accountService
}

type accountService interface {
	List(ctx context.Context, userID int64) ([]models.Account, *apperror.Error)
	Create(ctx context.Context, userID int64, req models.CreateAccountRequest) (*models.Account, *apperror.Error)
	GetByID(ctx context.Context, userID, accountID int64) (*models.Account, *apperror.Error)
	Update(ctx context.Context, userID, accountID int64, req models.UpdateAccountRequest) (*models.Account, *apperror.Error)
	Delete(ctx context.Context, userID, accountID int64) *apperror.Error
}

func NewAccountHandler(accountService *service.AccountService) *AccountHandler {
	if accountService == nil {
		return &AccountHandler{}
	}
	return &AccountHandler{accountService: accountService}
}

// List godoc
// @Summary List accounts
// @Description List all accounts for authenticated user.
// @Tags accounts
// @Security BearerAuth
// @Produce json
// @Success 200 {array} Account
// @Failure 401 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/accounts [get]
func (h *AccountHandler) List(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	out, appErr := h.accountService.List(c.Request.Context(), userID)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Create godoc
// @Summary Create account
// @Description Create an account for authenticated user.
// @Tags accounts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateAccountRequest true "Create account payload"
// @Success 201 {object} Account
// @Failure 400 {object} ErrorEnvelope
// @Failure 401 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/accounts [post]
func (h *AccountHandler) Create(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, apperror.Validation(err.Error()))
		return
	}
	out, appErr := h.accountService.Create(c.Request.Context(), userID, req)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusCreated, out)
}

// GetByID godoc
// @Summary Get account
// @Description Get account by id for authenticated user.
// @Tags accounts
// @Security BearerAuth
// @Produce json
// @Param id path int true "Account ID"
// @Success 200 {object} Account
// @Failure 400 {object} ErrorEnvelope
// @Failure 401 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Router /api/v1/accounts/{id} [get]
func (h *AccountHandler) GetByID(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(c, apperror.Validation("invalid account id"))
		return
	}
	out, appErr := h.accountService.GetByID(c.Request.Context(), userID, id)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Update godoc
// @Summary Update account
// @Description Update account fields for authenticated user.
// @Tags accounts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Account ID"
// @Param request body UpdateAccountRequest true "Update account payload"
// @Success 200 {object} Account
// @Failure 400 {object} ErrorEnvelope
// @Failure 401 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/accounts/{id} [patch]
func (h *AccountHandler) Update(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(c, apperror.Validation("invalid account id"))
		return
	}

	var req models.UpdateAccountRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		writeError(c, apperror.Validation(err.Error()))
		return
	}
	out, appErr := h.accountService.Update(c.Request.Context(), userID, id, req)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Delete godoc
// @Summary Delete account
// @Description Soft-delete account for authenticated user.
// @Tags accounts
// @Security BearerAuth
// @Produce json
// @Param id path int true "Account ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} ErrorEnvelope
// @Failure 401 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/accounts/{id} [delete]
func (h *AccountHandler) Delete(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(c, apperror.Validation("invalid account id"))
		return
	}
	if appErr := h.accountService.Delete(c.Request.Context(), userID, id); appErr != nil {
		writeError(c, appErr)
		return
	}
	c.Status(http.StatusNoContent)
}
