package controller

import (
	"net/http"
	"strconv"
	"time"
	"wx-purchase-api/model"
	"wx-purchase-api/usecase"

	"github.com/gin-gonic/gin"
)

type purchaseTransactionController struct {
	purchaseTransactionUsecase usecase.PurchaseTransactionUsecase
}

func (rc *purchaseTransactionController) HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, model.HealthResponse{
		Status:    "ok",
		Service:   "wx-purchase-api",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func NewPurchaseTransactionController(purchaseTransactionUsecase usecase.PurchaseTransactionUsecase) purchaseTransactionController {
	return purchaseTransactionController{
		purchaseTransactionUsecase,
	}
}

func (rc *purchaseTransactionController) GetTransactions(c *gin.Context) {
	ctx := c.Request.Context() // Use request context for better cancellation and timeout handling
	transactions, err := rc.purchaseTransactionUsecase.GetTransactions(ctx)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to retrieve transactions")
		return
	}

	c.JSON(http.StatusOK, transactions)
}

func (rc *purchaseTransactionController) GetTransactionById(c *gin.Context) {
	ctx := c.Request.Context() // Use request context for better cancellation and timeout handling
	idParam := c.Param("id")
	if idParam == "" {
		writeError(c, http.StatusBadRequest, "ID parameter is required")
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		writeError(c, http.StatusBadRequest, "Invalid ID parameter")
		return
	}

	transaction, err := rc.purchaseTransactionUsecase.GetTransactionById(ctx, id)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to retrieve transaction")

		return
	}

	c.JSON(http.StatusOK, transaction)
}

func (rc *purchaseTransactionController) CreateTransaction(c *gin.Context) {
	ctx := c.Request.Context() // Use request context for better cancellation and timeout handling
	var request struct {
		Description     string  `json:"description"`
		TransactionDate string  `json:"transactionDate"`
		Amount          float64 `json:"amount"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	transaction := model.PurchaseTransaction{
		Description:     request.Description,
		TransactionDate: request.TransactionDate,
		Amount:          request.Amount,
	}

	if err := rc.purchaseTransactionUsecase.SaveTransaction(ctx, transaction); err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to save transaction")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Transaction created successfully"})
}

func (rc *purchaseTransactionController) GetTransactionExchange(c *gin.Context) {
	ctx := c.Request.Context() // Use request context for better cancellation and timeout handling
	idParam := c.Param("id")
	if idParam == "" {
		writeError(c, http.StatusBadRequest, "ID parameter is required")
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		writeError(c, http.StatusBadRequest, "Invalid ID parameter")
		return
	}

	cParam := c.Param("currency")
	if cParam == "" {
		writeError(c, http.StatusBadRequest, "currency parameter is required")
		return
	}

	pte, err := rc.purchaseTransactionUsecase.GetTransactionExchange(ctx, id, cParam)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to retrieve transaction")

		return
	}

	c.JSON(http.StatusOK, pte)
}

func writeError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}
