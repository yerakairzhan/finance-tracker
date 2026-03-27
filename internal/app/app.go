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
	"context"
	"log"
	"os"
	"time"

	"finance-tracker/db/migrations"
	sqlc "finance-tracker/db/queries"
	_ "finance-tracker/docs"
	"finance-tracker/pkg/cache"
	"finance-tracker/pkg/handler"
	"finance-tracker/pkg/middleware"
	"finance-tracker/pkg/repository"
	"finance-tracker/pkg/service"

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
	redisDB := 0

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

	q := sqlc.New(pool)

	userRepo := repository.NewUserRepository(q)
	accountRepo := repository.NewAccountRepository(q)
	txRepo := repository.NewTransactionRepository(pool, q)

	authService := service.NewAuthService(userRepo, jwtSecret, tokenBlocklist)
	userService := service.NewUserService(userRepo)
	accountService := service.NewAccountService(accountRepo)
	txService := service.NewTransactionService(txRepo)
	healthService := service.NewHealthService(pool)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	accountHandler := handler.NewAccountHandler(accountService)
	transactionHandler := handler.NewTransactionHandler(txService)
	healthHandler := handler.NewHealthHandler(healthService)

	router := gin.Default()
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/health", healthHandler.Live)
	router.GET("/health/ready", healthHandler.Ready)

	v1 := router.Group("/api/v1")
	{
		authRoutes := v1.Group("/auth")
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
		authRoutes.POST("/refresh", authHandler.Refresh)

		protected := v1.Group("")
		protected.Use(middleware.JWTAuth(jwtSecret, tokenBlocklist))
		{
			authProtected := protected.Group("/auth")
			authProtected.POST("/logout", authHandler.Logout)

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
		}
	}

	log.Println("server running on port", port)
	if err = router.Run(":" + port); err != nil {
		log.Fatal("failed to start server: ", err)
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
