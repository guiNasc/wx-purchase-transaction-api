package main

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
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
	logger := buildLogger()
	slog.SetDefault(logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	router := gin.New()
	router.Use(gin.Recovery(), requestLoggerMiddleware(logger))

	dbConnection, err := database.ConnectDB()
	if err != nil {
		logger.Error("failed to connect database", "error", err)
		os.Exit(1)
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

	logger.Info("server starting", "port", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}

}

func buildLogger() *slog.Logger {
	level := parseLogLevel(os.Getenv("LOG_LEVEL"))
	format := strings.ToLower(os.Getenv("LOG_FORMAT"))

	options := &slog.HandlerOptions{Level: level}
	if format == "json" {
		return slog.New(slog.NewJSONHandler(os.Stdout, options))
	}

	return slog.New(slog.NewTextHandler(os.Stdout, options))
}

func parseLogLevel(value string) slog.Level {
	switch strings.ToLower(value) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func requestLoggerMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		logger.Info("http_request",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}
