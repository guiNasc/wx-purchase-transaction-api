package usecase

import (
	"context"
	"fmt"
	"time"
	"wx-purchase-api/model"
)

const maxDescriptionLength = 50

type IPurchaseRepository interface {
	Get() ([]model.PurchaseTransaction, error)
	Save(transaction model.PurchaseTransaction) error
	GetById(id int) (model.PurchaseTransaction, error)
}

type IRequestGateway interface {
	GetExchangeRate(ctx context.Context, endpoint string, currency string) (model.ExchangeAPIResponse, error)
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

	purchaseDate, err := time.Parse(time.RFC3339, p.TransactionDate)
	if err != nil {
		return model.PurchaseTransactionExchange{}, fmt.Errorf("invalid transaction date format: %w", err)
	}

	sixMonthsAgo := purchaseDate.AddDate(0, -6, 0).Format("2006-01-02")

	baseUrl := "https://api.fiscaldata.treasury.gov/services/api/fiscal_service/"
	endpoint := fmt.Sprintf("%sv1/accounting/od/rates_of_exchange", baseUrl)
	url := fmt.Sprintf("%s?fields=record_date,country,currency,country_currency_desc,exchange_rate", endpoint)
	url += fmt.Sprintf("&filter=currency:in:(%s),record_date:gte:%s,record_date:lte:%s&sort=-record_date", currency, sixMonthsAgo,
		purchaseDate.Format("2006-01-02"),
	)

	fmt.Println(url)

	apiResponse, err := ptu.requestGateway.GetExchangeRate(ctx, url, currency)
	if err != nil {
		return model.PurchaseTransactionExchange{}, err
	}

	for _, exchangeData := range apiResponse.Data {
		fmt.Println(exchangeData)
	}

	return model.PurchaseTransactionExchange{}, nil
}
