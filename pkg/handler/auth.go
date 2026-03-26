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

type authService interface {
	Register(ctx context.Context, req models.RegisterRequest) (*models.AuthTokens, *apperror.Error)
	Login(ctx context.Context, req models.LoginRequest) (*models.AuthTokens, *apperror.Error)
	Refresh(ctx context.Context, rawRefreshToken string) (*models.AuthTokens, *apperror.Error)
	Logout(ctx context.Context, userID int64, rawRefreshToken string) *apperror.Error
}

type AuthHandler struct {
	authService authService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary Register
// @Description Register a new user and return access/refresh tokens.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register payload"
// @Success 201 {object} AuthTokens
// @Failure 400 {object} ErrorEnvelope
// @Failure 409 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, apperror.Validation(err.Error()))
		return
	}
	out, appErr := h.authService.Register(c.Request.Context(), req)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusCreated, out)
}

// Login godoc
// @Summary Login
// @Description Login with email and password.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login payload"
// @Success 200 {object} AuthTokens
// @Failure 400 {object} ErrorEnvelope
// @Failure 401 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, apperror.Validation(err.Error()))
		return
	}
	out, appErr := h.authService.Login(c.Request.Context(), req)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Refresh godoc
// @Summary Refresh tokens
// @Description Rotate refresh token and return new tokens.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh payload"
// @Success 200 {object} AuthTokens
// @Failure 400 {object} ErrorEnvelope
// @Failure 401 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req models.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, apperror.Validation(err.Error()))
		return
	}
	out, appErr := h.authService.Refresh(c.Request.Context(), req.RefreshToken)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Logout godoc
// @Summary Logout
// @Description Revoke current refresh token.
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "Logout payload"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} ErrorEnvelope
// @Failure 401 {object} ErrorEnvelope
// @Failure 404 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req models.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, apperror.Validation(err.Error()))
		return
	}
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	if appErr := h.authService.Logout(c.Request.Context(), userID, req.RefreshToken); appErr != nil {
		writeError(c, appErr)
		return
	}
	c.Status(http.StatusNoContent)
}
