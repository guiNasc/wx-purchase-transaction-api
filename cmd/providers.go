package main

import (
	"database/sql"
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
)

const defaultPort = "8080"

func NewLogger() *slog.Logger {
	level := parseLogLevel(os.Getenv("LOG_LEVEL"))
	format := strings.ToLower(os.Getenv("LOG_FORMAT"))

	options := &slog.HandlerOptions{Level: level}
	if format == "json" {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, options))
		slog.SetDefault(logger)
		return logger
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, options))
	slog.SetDefault(logger)
	return logger
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

func ProvidePort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return defaultPort
	}

	return port
}

func ProvideDB(_ *slog.Logger) (*sql.DB, func(), error) {
	db, err := database.ConnectDB()
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		if err := db.Close(); err != nil {
			slog.Default().Error("failed to close database connection", "error", err)
		}
	}

	return db, cleanup, nil
}

func NewRouter(transactionUsecase usecase.PurchaseTransactionUsecase, logger *slog.Logger) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery(), requestLoggerMiddleware(logger))

	transactionController := controller.NewPurchaseTransactionController(transactionUsecase)
	router.GET("/health", transactionController.HealthHandler)
	router.GET("/transactions", transactionController.GetTransactions)
	router.POST("/transactions", transactionController.CreateTransaction)
	router.GET("/transactions/:id", transactionController.GetTransactionById)
	router.GET("/transactions/:id/exchange/:currency", transactionController.GetTransactionExchange)

	return router
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

func NewHTTPServer(router *gin.Engine, port string) *http.Server {
	return &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func NewTransactionUsecase(repo usecase.IPurchaseRepository, gateway usecase.IRequestGateway) usecase.PurchaseTransactionUsecase {
	return usecase.NewPurchaseTransactionUseCase(repo, gateway)
}

func NewPurchaseRepository(db *sql.DB) usecase.IPurchaseRepository {
	return repository.NewPurchaseTransactionRepository(db)
}

func NewRequestGateway() usecase.IRequestGateway {
	return infra.NewGateway()
}
