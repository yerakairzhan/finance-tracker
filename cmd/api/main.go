package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"

	"finance-tracker/pkg/generated/sqlc"
	"finance-tracker/pkg/handler"
	"finance-tracker/pkg/repository"
)

func main() {

	// DATABASE CONNECTION STRING
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/financial_intelligence?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Verify connection
	if err := db.PingContext(ctx); err != nil {
		log.Fatal("Database unreachable:", err)
	}

	log.Println("Connected to PostgreSQL")

	queries := sqlc.New(db) // ✅ db implements DBTX

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

		// Transactions
		api.GET("/transactions", transactionHandler.List)
	}

	// START SERVER
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server running on port", port)
	router.Run(":" + port)
}