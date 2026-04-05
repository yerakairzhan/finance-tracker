package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func DoubleSubmitCSRF(required bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !required {
			c.Next()
			return
		}
		header := strings.TrimSpace(c.GetHeader("X-CSRF-Token"))
		cookie, err := c.Cookie("csrf_token")
		if err != nil || header == "" || cookie == "" || header != cookie {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "UNAUTHORIZED", "message": "csrf validation failed"},
			})
			return
		}
		c.Next()
	}
}
