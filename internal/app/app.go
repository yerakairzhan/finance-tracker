// @title Finance Tracker API
// @version 1.0
// @description Personal finance tracking API (v1 core).
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Use value: Bearer <access_token>
package app

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"finance-tracker/db/migrations"
	sqlc "finance-tracker/db/queries"
	_ "finance-tracker/docs"
	"finance-tracker/pkg/cache"
	"finance-tracker/pkg/handler"
	"finance-tracker/pkg/middleware"
	"finance-tracker/pkg/repository"
	"finance-tracker/pkg/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Run() {
	dbURL := getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/finance_tracker?sslmode=disable")
	port := getenv("PORT", "8080")
	jwtSecret := getenv("JWT_SECRET", "dev-jwt-secret-change-me")
	redisAddr := getenv("REDIS_ADDR", "localhost:6379")
	redisPassword := getenv("REDIS_PASSWORD", "")
	frontendOrigin := getenv("FRONTEND_ORIGIN", "http://localhost:5173")
	cookieSameSite := parseSameSite(getenv("COOKIE_SAMESITE", "strict"))
	refreshPepper := getenv("REFRESH_TOKEN_PEPPER", "dev-refresh-pepper-change-me")
	redisDB := 0
	accessTokenResponseMode := strings.ToLower(strings.TrimSpace(getenv("ACCESS_TOKEN_RESPONSE_MODE", "omit"))) // hashed|omit
	if accessTokenResponseMode != "hashed" && accessTokenResponseMode != "omit" {
		accessTokenResponseMode = "omit"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal("failed to connect to database: ", err)
	}
	if err = pool.Ping(ctx); err != nil {
		log.Fatal("database unreachable: ", err)
	}
	log.Println("connected to PostgreSQL")

	migrationCtx, migrationCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer migrationCancel()

	if err = migrations.Run(migrationCtx, pool); err != nil {
		log.Fatal("failed to apply migrations: ", err)
	}

	redisClient := cache.NewRedisClient(redisAddr, redisPassword, redisDB)
	if err = redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal("failed to connect to redis: ", err)
	}
	log.Println("connected to Redis")
	tokenBlocklist := cache.NewTokenBlocklist(redisClient)
	refreshStore := cache.NewRefreshSessionStore(redisClient)

	q := sqlc.New(pool)

	userRepo := repository.NewUserRepository(q)
	accountRepo := repository.NewAccountRepository(q)
	txRepo := repository.NewTransactionRepository(pool, q)

	authService := service.NewAuthService(
		userRepo,
		jwtSecret,
		tokenBlocklist,
		refreshStore,
		refreshPepper,
	)
	userService := service.NewUserService(userRepo)
	accountService := service.NewAccountService(accountRepo)
	txService := service.NewTransactionService(txRepo)
	analyticsService := service.NewAnalyticsService(txRepo)
	healthService := service.NewHealthService(pool)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	accountHandler := handler.NewAccountHandler(accountService)
	transactionHandler := handler.NewTransactionHandler(txService)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService)
	healthHandler := handler.NewHealthHandler(healthService)

	authRateLimiter := middleware.NewAuthRateLimiter(redisClient, middleware.AuthRateLimitConfig{
		LoginLimit:    10,
		LoginWindow:   5 * time.Minute,
		RefreshLimit:  30,
		RefreshWindow: 5 * time.Minute,
	})
	csrfRequired := cookieSameSite == http.SameSiteNoneMode

	router := gin.Default()
	router.Use(sanitizeAuthResponseMiddleware()) // no token hashing, only sensitive field stripping
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendOrigin},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/health", healthHandler.Live)
	router.GET("/health/ready", healthHandler.Ready)

	v1 := router.Group("/api/v1")
	{
		authRoutes := v1.Group("/auth")
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authRateLimiter.LoginLimiter(), authHandler.Login)
		authRoutes.POST("/refresh", authRateLimiter.RefreshLimiter(), middleware.DoubleSubmitCSRF(csrfRequired), authHandler.Refresh)

		protected := v1.Group("")
		protected.Use(middleware.JWTAuth(jwtSecret, tokenBlocklist))
		{
			authProtected := protected.Group("/auth")
			authProtected.POST("/logout", middleware.DoubleSubmitCSRF(csrfRequired), authHandler.Logout)

			userRoutes := protected.Group("/users")
			userRoutes.GET("/me", userHandler.Me)
			userRoutes.PATCH("/me", userHandler.UpdateMe)
			userRoutes.PATCH("/me/password", userHandler.ChangePassword)

			accountRoutes := protected.Group("/accounts")
			accountRoutes.GET("", accountHandler.List)
			accountRoutes.POST("", accountHandler.Create)
			accountRoutes.GET("/:id", accountHandler.GetByID)
			accountRoutes.PATCH("/:id", accountHandler.Update)
			accountRoutes.DELETE("/:id", accountHandler.Delete)

			txRoutes := protected.Group("/transactions")
			txRoutes.GET("", transactionHandler.List)
			txRoutes.POST("", transactionHandler.Create)
			txRoutes.GET("/:id", transactionHandler.GetByID)
			txRoutes.PATCH("/:id", transactionHandler.Update)
			txRoutes.DELETE("/:id", transactionHandler.Delete)

			analyticsRoutes := protected.Group("/analytics")
			analyticsRoutes.GET("/summary/last-month", analyticsHandler.LastMonthSummary)
			analyticsRoutes.GET("/daily-profit", analyticsHandler.DailyProfit)
			analyticsRoutes.GET("/expense-categories/last-month", analyticsHandler.LastMonthExpenseByCategory)
			analyticsRoutes.GET("/monthly-profit", analyticsHandler.MonthlyProfit)
		}
	}

	log.Println("server running on port", port)
	if err = router.Run(":" + port); err != nil {
		log.Fatal("failed to start server: ", err)
	}
}

func parseSameSite(v string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "none":
		return http.SameSiteNoneMode
	case "lax":
		return http.SameSiteLaxMode
	default:
		return http.SameSiteStrictMode
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

type responseCaptureWriter struct {
	gin.ResponseWriter
	body       bytes.Buffer
	statusCode int
}

func (w *responseCaptureWriter) WriteHeader(code int) { w.statusCode = code }
func (w *responseCaptureWriter) WriteHeaderNow()      {}
func (w *responseCaptureWriter) Write(data []byte) (int, error) {
	return w.body.Write(data)
}
func (w *responseCaptureWriter) WriteString(s string) (int, error) {
	return w.body.WriteString(s)
}

func sanitizeAuthResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAuthTokenResponsePath(c.Request.URL.Path) {
			c.Next()
			return
		}

		original := c.Writer
		capture := &responseCaptureWriter{ResponseWriter: original, statusCode: http.StatusOK}
		c.Writer = capture
		c.Next()

		status := capture.statusCode
		if status == 0 {
			status = original.Status()
			if status == 0 {
				status = http.StatusOK
			}
		}

		sanitized := sanitizeTokenFields(capture.body.Bytes())
		original.Header().Del("Content-Length")
		original.WriteHeader(status)
		_, _ = original.Write(sanitized)
	}
}

func isAuthTokenResponsePath(path string) bool {
	p := strings.TrimSpace(path)
	if len(p) > 1 {
		p = strings.TrimRight(p, "/")
	}
	switch p {
	case "/api/v1/auth/register", "/api/v1/auth/login", "/api/v1/auth/refresh":
		return true
	default:
		return false
	}
}

func sanitizeTokenFields(body []byte) []byte {
	if len(body) == 0 {
		return body
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return body
	}

	changed := false

	// Never expose refresh/session internals in JSON
	for _, k := range []string{"refresh_token", "refresh_session_id", "session_id", "jti", "token_id"} {
		if _, ok := payload[k]; ok {
			delete(payload, k)
			changed = true
		}
	}

	if !changed {
		return body
	}
	out, err := json.Marshal(payload)
	if err != nil {
		return body
	}
	return out
}