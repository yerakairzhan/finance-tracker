// @title Finance Tracker API
// @version 1.0
// @description API for managing users, accounts, and transactions
// @host localhost:8080
// @BasePath /

package app

import (
	"context"
	"log"
	"os"
	"time"
	"github.com/gin-gonic/gin"

	"github.com/jackc/pgx/v5/pgxpool"

	"finance-tracker/pkg/generated/sqlc"
	"finance-tracker/pkg/handler"
	"finance-tracker/pkg/repository"
)

func Run() {

	// DATABASE CONNECTION STRING
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/financial_intelligence?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("Database unreachable:", err)
	}

	log.Println("Connected to PostgreSQL")

	queries := sqlc.New(pool)

	// REPOSITORIES
	userRepo := repository.NewUserRepository(queries)
	accountRepo := repository.NewAccountRepository(queries)
	transactionRepo := repository.NewTransactionRepository(queries)

	// HANDLERs
	userHandler := handler.NewUserHandler(userRepo)
	accountHandler := handler.NewAccountHandler(accountRepo)
	transactionHandler := handler.NewTransactionHandler(transactionRepo)

	// GIN ROUTER
	router := gin.Default()

	// API ROUTES
	api := router.Group("")
	{
		// User
		api.POST("/register", userHandler.Register)

		api.GET("/users", userHandler.List)

		api.GET("/users/:id", userHandler.GetByID)

		api.PUT("/users/:id", userHandler.Update)

		api.DELETE("/users/:id", userHandler.Delete)

		// Accounts
		api.POST("/accounts", accountHandler.Create)

		api.GET("/accounts", accountHandler.List)

		api.GET("/accounts/:id", accountHandler.GetByID)

		api.GET("/users/:id/accounts", accountHandler.GetUserAccounts)

		api.DELETE("/accounts/:id", accountHandler.Delete)

		api.GET("/accounts/:id/balance", accountHandler.GetBalance)

		// Transactions
		api.POST("/transactions", transactionHandler.Create)

		api.GET("/transactions", transactionHandler.List)

		api.GET("/transactions/:id", transactionHandler.GetByID)

		api.GET("/accounts/:id/transactions", transactionHandler.GetByAccount)

		api.DELETE("/transactions/:id", transactionHandler.Delete)

		api.GET("/transactions/search", transactionHandler.Search)

		api.GET("/transactions/export", transactionHandler.Export)
	}

	// START SERVER
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server running on port", port)
	router.Run(":" + port)
}