package main

import (
	"log"
	"net/http"
	"os"
	"time"
	"wx-purchase-api/controller"
	"wx-purchase-api/database"
	"wx-purchase-api/infra"
	"wx-purchase-api/repository"
	"wx-purchase-api/usecase"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const defaultPort = "8080"

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	dbConnection, err := database.ConnectDB()
	if err != nil {
		panic(err)
	}

	purchaseTransactionRepository := repository.NewPurchaseTransactionRepository(dbConnection)

	gateway := infra.NewGateway()
	purchaseTransactionUsecase := usecase.NewPurchaseTransactionUseCase(purchaseTransactionRepository, gateway)
	purchaseTransactionController := controller.NewPurchaseTransactionController(purchaseTransactionUsecase)

	router.GET("/health", purchaseTransactionController.HealthHandler)
	router.GET("/transactions", purchaseTransactionController.GetTransactions)
	router.POST("/transactions", purchaseTransactionController.CreateTransaction)
	router.GET("/transactions/:id", purchaseTransactionController.GetTransactionById)
	router.GET("/transactions/:id/exchange/:currency", purchaseTransactionController.GetTransactionExchange)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("wx-purchase-api listening on :%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}

}
