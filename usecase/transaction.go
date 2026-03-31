package usecase

import (
	"context"
	"strings"
	"time"
	"wx-purchase-api/apperror"
	"wx-purchase-api/model"

	"github.com/shopspring/decimal"
)

const maxDescriptionLength = 50

type IPurchaseRepository interface {
	Get(ctx context.Context) ([]model.PurchaseTransaction, error)
	Save(ctx context.Context, transaction model.PurchaseTransaction) (model.PurchaseTransaction, error)
	GetById(ctx context.Context, id int) (model.PurchaseTransaction, error)
}

type IRequestGateway interface {
	GetExchangeRate(ctx context.Context, currency, fromDate, toDate string) (model.ExchangeAPIResponse, error)
}

type PurchaseTransactionUsecase struct {
	purchaseTransactionRepository IPurchaseRepository
	requestGateway                IRequestGateway
}

func NewPurchaseTransactionUseCase(purchaseTransactionRepository IPurchaseRepository, requestGateway IRequestGateway) PurchaseTransactionUsecase {
	return PurchaseTransactionUsecase{
		purchaseTransactionRepository,
		requestGateway,
	}
}

func (ptu *PurchaseTransactionUsecase) GetTransactions(ctx context.Context) ([]model.PurchaseTransaction, error) {
	return ptu.purchaseTransactionRepository.Get(ctx)
}

func (ptu *PurchaseTransactionUsecase) SaveTransaction(ctx context.Context, transaction model.PurchaseTransaction) (model.PurchaseTransaction, error) {
	if err := validateTransaction(transaction); err != nil {
		return model.PurchaseTransaction{}, err
	}

	return ptu.purchaseTransactionRepository.Save(ctx, transaction)
}

func validateTransaction(transaction model.PurchaseTransaction) error {
	if len(transaction.Description) > maxDescriptionLength {
		return apperror.Unprocessable("description_too_long", "description must not be longer than 50 characters", nil)
	}

	if transaction.Amount <= 0 {
		return apperror.Unprocessable("invalid_amount", "amount must be positive", nil)
	}

	if _, err := time.Parse("2006-01-02", transaction.TransactionDate); err != nil {
		return apperror.Unprocessable("invalid_transaction_date", "transactionDate must be in YYYY-MM-DD format", err)
	}

	return nil
}

func (ptu *PurchaseTransactionUsecase) GetTransactionById(ctx context.Context, id int) (model.PurchaseTransaction, error) {
	if id <= 0 {
		return model.PurchaseTransaction{}, apperror.Unprocessable("invalid_id", "id must be greater than zero", nil)
	}

	return ptu.purchaseTransactionRepository.GetById(ctx, id)
}

func (ptu *PurchaseTransactionUsecase) GetTransactionExchange(ctx context.Context, id int, currency string) (model.PurchaseTransactionExchange, error) {
	if id <= 0 {
		return model.PurchaseTransactionExchange{}, apperror.Unprocessable("invalid_id", "id must be greater than zero", nil)
	}

	if strings.TrimSpace(currency) == "" {
		return model.PurchaseTransactionExchange{}, apperror.BadRequest("missing_currency", "currency is required", nil)
	}

	p, err := ptu.purchaseTransactionRepository.GetById(ctx, id)
	if err != nil {
		return model.PurchaseTransactionExchange{}, err
	}

	toDate, err := time.Parse("2006-01-02", p.TransactionDate)
	if err != nil {
		return model.PurchaseTransactionExchange{}, apperror.Unprocessable("invalid_transaction_date", "transaction has invalid date format", err)
	}

	fromDate := toDate.AddDate(0, -6, 0).Format("2006-01-02")
	toDateStr := toDate.Format("2006-01-02")

	apiResponse, err := ptu.requestGateway.GetExchangeRate(ctx, currency, fromDate, toDateStr)
	if err != nil {
		return model.PurchaseTransactionExchange{}, err
	}

	if apiResponse.Meta.Count == 0 {
		return model.PurchaseTransactionExchange{}, apperror.Unprocessable("exchange_rate_not_found", "no exchange rate data found for the requested period", nil)
	}

	apiObj := apiResponse.Data[0]

	eRate, err := decimal.NewFromString(apiObj.ExchangeRate)
	if err != nil {
		return model.PurchaseTransactionExchange{}, apperror.ServiceUnavailable("invalid_exchange_rate_response", "exchange service returned invalid data", err)
	}

	decimalAmount := decimal.NewFromFloat(p.Amount)

	convertedAmount := decimalAmount.Mul(eRate).Round(2)

	return model.PurchaseTransactionExchange{
		Description:     p.Description,
		TransactionDate: p.TransactionDate,
		Amount:          p.Amount,
		ID:              p.ID,
		Currency:        apiObj.Currency,
		ExchangeRate:    eRate.InexactFloat64(),
		ConvertedAmount: convertedAmount.InexactFloat64(),
	}, nil
}
