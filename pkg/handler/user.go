package handler

import (
	"net/http"

	"github.com/example/financial-intelligence-platform/pkg/models"
	"github.com/example/financial-intelligence-platform/pkg/repository"
	"github.com/gin-gonic/gin"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	repo *repository.UserRepository
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(repo *repository.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

// Register creates a new user
// POST /register
func (h *UserHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.repo.CreateUser(c.Request.Context(), req.Email, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fail"})
		return
	}

	c.JSON(http.StatusCreated, user)
}
