package usecase

import (
	"context"
	"fmt"
	"time"
	"wx-purchase-api/model"

	"github.com/shopspring/decimal"
)

const maxDescriptionLength = 50

type IPurchaseRepository interface {
	Get() ([]model.PurchaseTransaction, error)
	Save(transaction model.PurchaseTransaction) error
	GetById(id int) (model.PurchaseTransaction, error)
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

func (ptu *PurchaseTransactionUsecase) GetTransactions() ([]model.PurchaseTransaction, error) {
	return ptu.purchaseTransactionRepository.Get()
}

func (ptu *PurchaseTransactionUsecase) SaveTransaction(transaction model.PurchaseTransaction) error {
	if err := validateTransaction(transaction); err != nil {
		return err
	}

	return ptu.purchaseTransactionRepository.Save(transaction)
}

func validateTransaction(transaction model.PurchaseTransaction) error {
	if len(transaction.Description) > maxDescriptionLength {
		return fmt.Errorf("description must not be longer than %d characters", maxDescriptionLength)
	}

	if transaction.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	if _, err := time.Parse("2006-01-02", transaction.TransactionDate); err != nil {
		return fmt.Errorf("transactionDate must be in YYYY-MM-DD format")
	}

	return nil
}

func (ptu *PurchaseTransactionUsecase) GetTransactionById(id int) (model.PurchaseTransaction, error) {
	return ptu.purchaseTransactionRepository.GetById(id)
}

func (ptu *PurchaseTransactionUsecase) GetTransactionExchange(ctx context.Context, id int, currency string) (model.PurchaseTransactionExchange, error) {
	p, err := ptu.purchaseTransactionRepository.GetById(id)
	if err != nil {
		return model.PurchaseTransactionExchange{}, err
	}

	toDate, err := time.Parse(time.RFC3339, p.TransactionDate)
	if err != nil {
		return model.PurchaseTransactionExchange{}, fmt.Errorf("invalid transaction date format: %w", err)
	}

	fromDate := toDate.AddDate(0, -6, 0).Format("2006-01-02")
	toDateStr := toDate.Format("2006-01-02")

	apiResponse, err := ptu.requestGateway.GetExchangeRate(ctx, currency, fromDate, toDateStr)
	if err != nil {
		return model.PurchaseTransactionExchange{}, err
	}

	if apiResponse.Meta.Count == 0 {
		return model.PurchaseTransactionExchange{}, fmt.Errorf("no exchange rate data found for currency %s in the last 6 months", currency)
	}

	apiObj := apiResponse.Data[0]

	eRate, err := decimal.NewFromString(apiObj.ExchangeRate)
	if err != nil {
		return model.PurchaseTransactionExchange{}, fmt.Errorf("invalid exchange rate format: %w", err)
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
