package middleware

import (
	"context"
	"strings"

	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/auth"

	"github.com/gin-gonic/gin"
)

const userIDContextKey = "auth_user_id"

type tokenBlocklist interface {
	IsRevoked(ctx context.Context, tokenID string) (bool, error)
}

func UserIDFromContext(c *gin.Context) (int64, bool) {
	v, ok := c.Get(userIDContextKey)
	if !ok {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}

func AccessTokenFromHeader(authz string) (string, *apperror.Error) {
	if authz == "" {
		return "", apperror.Unauthorized("missing bearer token")
	}
	parts := strings.SplitN(authz, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", apperror.Unauthorized("invalid authorization header")
	}
	return strings.TrimSpace(parts[1]), nil
}

func JWTAuth(secret string, blocklist tokenBlocklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawToken, authErr := AccessTokenFromHeader(c.GetHeader("Authorization"))
		if authErr != nil {
			writeAbort(c, authErr)
			return
		}
		claims, err := auth.ParseAccessToken(secret, rawToken)
		if err != nil {
			writeAbort(c, apperror.Unauthorized("invalid or expired token"))
			return
		}
		if claims.ID == "" {
			writeAbort(c, apperror.Unauthorized("invalid token"))
			return
		}
		revoked, err := blocklist.IsRevoked(c.Request.Context(), claims.ID)
		if err != nil {
			writeAbort(c, apperror.Internal("failed to verify token"))
			return
		}
		if revoked {
			writeAbort(c, apperror.Unauthorized("token has been revoked"))
			return
		}
		c.Set(userIDContextKey, claims.UserID)
		c.Next()
	}
}

func writeAbort(c *gin.Context, err *apperror.Error) {
	c.AbortWithStatusJSON(err.Status, gin.H{
		"error": gin.H{
			"code":    err.Code,
			"message": err.Message,
		},
	})
}
