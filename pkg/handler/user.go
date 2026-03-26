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

type UserHandler struct {
	userService userService
}

type userService interface {
	Me(ctx context.Context, userID int64) (*models.User, *apperror.Error)
	UpdateMe(ctx context.Context, userID int64, req models.UpdateMeRequest) (*models.User, *apperror.Error)
	ChangePassword(ctx context.Context, userID int64, req models.ChangePasswordRequest) *apperror.Error
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	if userService == nil {
		return &UserHandler{}
	}
	return &UserHandler{userService: userService}
}

// Me godoc
// @Summary Get profile
// @Description Get authenticated user's profile.
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} User
// @Failure 401 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Router /api/v1/users/me [get]
func (h *UserHandler) Me(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	out, appErr := h.userService.Me(c.Request.Context(), userID)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, out)
}

// UpdateMe godoc
// @Summary Update profile
// @Description Update authenticated user's name/currency.
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpdateMeRequest true "Update profile payload"
// @Success 200 {object} User
// @Failure 400 {object} ErrorEnvelope
// @Failure 401 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/users/me [patch]
func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}

	var req models.UpdateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, apperror.Validation(err.Error()))
		return
	}

	out, appErr := h.userService.UpdateMe(c.Request.Context(), userID, req)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, out)
}

// ChangePassword godoc
// @Summary Change password
// @Description Change authenticated user's password.
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "Change password payload"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} ErrorEnvelope
// @Failure 401 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/users/me/password [patch]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, apperror.Validation(err.Error()))
		return
	}
	if appErr := h.userService.ChangePassword(c.Request.Context(), userID, req); appErr != nil {
		writeError(c, appErr)
		return
	}
	c.Status(http.StatusNoContent)
}
