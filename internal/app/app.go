package app

import (
	"fmt"
	"log"
	"os"

	"finance-tracker/config"
	"finance-tracker/pkg/generated/sqlc"
	"finance-tracker/pkg/handler"
	"finance-tracker/pkg/repository"
	"github.com/gin-gonic/gin"
)

func Run() {
	dbHost := getenv("DB_HOST", "localhost")
	dbPort := getenv("DB_PORT", "5432")
	dbUser := getenv("DB_USER", "postgres")
	dbPassword := getenv("DB_PASSWORD", "postgres")
	dbName := getenv("DB_NAME", "financial_intelligence")

	dbConfig := config.DatabaseConfig{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		DBName:   dbName,
		SSLMode:  "disable",
	}

	db, err := config.NewDatabase(dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("Connected to database")

	queries := sqlc.New(db)

	userRepo := repository.NewUserRepository(queries)
	accountRepo := repository.NewAccountRepository(queries)
	transactionRepo := repository.NewTransactionRepository(queries)

	userHandler := handler.NewUserHandler(userRepo)
	accountHandler := handler.NewAccountHandler(accountRepo)
	transactionHandler := handler.NewTransactionHandler(transactionRepo)

	router := gin.Default()
	router.POST("/register", userHandler.Register)
	router.POST("/accounts", accountHandler.Create)
	router.GET("/transactions", transactionHandler.List)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := getenv("PORT", "8080")
	fmt.Printf("Server running on :%s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
