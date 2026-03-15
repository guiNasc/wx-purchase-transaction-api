package controller

import (
	"net/http"
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

	transactions, err := rc.purchaseTransactionUsecase.GetTransactions()
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to retrieve transactions")
		return
	}

	c.JSON(http.StatusOK, transactions)
}

func (rc *purchaseTransactionController) CreateTransaction(c *gin.Context) {
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

	if err := rc.purchaseTransactionUsecase.SaveTransaction(transaction); err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to save transaction")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Transaction created successfully"})
}

func writeError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}
