package handler

import (
	"context"
	"net/http"

	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/service"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	healthService healthService
}

type healthService interface {
	Ready(ctx context.Context) *apperror.Error
}

func NewHealthHandler(healthService *service.HealthService) *HealthHandler {
	if healthService == nil {
		return &HealthHandler{}
	}
	return &HealthHandler{healthService: healthService}
}

// Live godoc
// @Summary Liveness check
// @Description Returns 200 when process is alive.
// @Tags health
// @Produce json
// @Success 200 {object} StatusResponse
// @Router /health [get]
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Ready godoc
// @Summary Readiness check
// @Description Returns 200 when database is reachable.
// @Tags health
// @Produce json
// @Success 200 {object} StatusResponse
// @Failure 500 {object} ErrorEnvelope
// @Router /health/ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	if err := h.healthService.Ready(c.Request.Context()); err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}
