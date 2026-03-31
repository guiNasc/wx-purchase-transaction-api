package controller

import (
	"errors"
	"net/http"
	"strconv"
	"time"
	"wx-purchase-api/apperror"
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
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, transactions)
}

func (rc *purchaseTransactionController) GetTransactionById(c *gin.Context) {
	ctx := c.Request.Context() // Use request context for better cancellation and timeout handling
	idParam := c.Param("id")
	if idParam == "" {
		writeError(c, apperror.BadRequest("missing_id", "id parameter is required", nil))
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		writeError(c, apperror.BadRequest("invalid_id", "id parameter must be a valid integer", err))
		return
	}

	if id <= 0 {
		writeError(c, apperror.Unprocessable("invalid_id", "id parameter must be greater than zero", nil))
		return
	}

	transaction, err := rc.purchaseTransactionUsecase.GetTransactionById(ctx, id)
	if err != nil {
		writeError(c, err)
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
		writeError(c, apperror.BadRequest("invalid_request_body", "request body is invalid", err))
		return
	}

	transaction := model.PurchaseTransaction{
		Description:     request.Description,
		TransactionDate: request.TransactionDate,
		Amount:          request.Amount,
	}

	saved, err := rc.purchaseTransactionUsecase.SaveTransaction(ctx, transaction)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Transaction created successfully", "transaction": saved})
}

func (rc *purchaseTransactionController) GetTransactionExchange(c *gin.Context) {
	ctx := c.Request.Context() // Use request context for better cancellation and timeout handling
	idParam := c.Param("id")
	if idParam == "" {
		writeError(c, apperror.BadRequest("missing_id", "id parameter is required", nil))
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		writeError(c, apperror.BadRequest("invalid_id", "id parameter must be a valid integer", err))
		return
	}

	if id <= 0 {
		writeError(c, apperror.Unprocessable("invalid_id", "id parameter must be greater than zero", nil))
		return
	}

	cParam := c.Param("currency")
	if cParam == "" {
		writeError(c, apperror.BadRequest("missing_currency", "currency parameter is required", nil))
		return
	}

	pte, err := rc.purchaseTransactionUsecase.GetTransactionExchange(ctx, id, cParam)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, pte)
}

func writeError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	response := model.APIErrorResponse{
		Error: model.APIErrorBody{
			Code:    "internal_error",
			Message: "internal server error",
		},
	}

	var domainErr *apperror.Error
	if errors.As(err, &domainErr) {
		switch domainErr.Kind {
		case apperror.KindBadRequest:
			status = http.StatusBadRequest
		case apperror.KindNotFound:
			status = http.StatusNotFound
		case apperror.KindConflict:
			status = http.StatusConflict
		case apperror.KindUnprocessable:
			status = http.StatusUnprocessableEntity
		case apperror.KindRateLimited:
			status = http.StatusTooManyRequests
		case apperror.KindServiceUnavailable:
			status = http.StatusServiceUnavailable
		}

		response.Error.Code = domainErr.Code
		response.Error.Message = domainErr.Message
	}

	c.JSON(status, response)
}
