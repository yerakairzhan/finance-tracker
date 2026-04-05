package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type AuthRateLimitConfig struct {
	LoginLimit   int64
	LoginWindow  time.Duration
	RefreshLimit int64
	RefreshWindow time.Duration
}

type AuthRateLimiter struct {
	client *redis.Client
	cfg    AuthRateLimitConfig
}

func NewAuthRateLimiter(client *redis.Client, cfg AuthRateLimitConfig) *AuthRateLimiter {
	return &AuthRateLimiter{client: client, cfg: cfg}
}

func (r *AuthRateLimiter) LoginLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		if r == nil || r.client == nil {
			c.Next()
			return
		}
		ip := c.ClientIP()
		if blocked(c, r.client, "rl:auth:login:ip:"+ip, r.cfg.LoginLimit, r.cfg.LoginWindow) {
			return
		}

		var body struct {
			Email string `json:"email"`
		}
		raw, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(raw))
		_ = json.Unmarshal(raw, &body)

		email := strings.ToLower(strings.TrimSpace(body.Email))
		if email != "" && blocked(c, r.client, "rl:auth:login:email:"+email, r.cfg.LoginLimit, r.cfg.LoginWindow) {
			return
		}
		c.Next()
	}
}

func (r *AuthRateLimiter) RefreshLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		if r == nil || r.client == nil {
			c.Next()
			return
		}
		ip := c.ClientIP()
		if blocked(c, r.client, "rl:auth:refresh:ip:"+ip, r.cfg.RefreshLimit, r.cfg.RefreshWindow) {
			return
		}
		c.Next()
	}
}

func blocked(c *gin.Context, client *redis.Client, key string, limit int64, window time.Duration) bool {
	if limit <= 0 || window <= 0 {
		return false
	}
	pipe := client.TxPipeline()
	incr := pipe.Incr(c.Request.Context(), key)
	pipe.Expire(c.Request.Context(), key, window)
	_, err := pipe.Exec(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": gin.H{"code": "RATE_LIMITED", "message": "too many requests"}})
		return true
	}
	if incr.Val() > limit {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": gin.H{"code": "RATE_LIMITED", "message": "too many requests"}})
		return true
	}
	return false
}
