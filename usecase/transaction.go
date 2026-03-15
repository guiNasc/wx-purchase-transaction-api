package usecase

import (
	"fmt"
	"time"
	"wx-purchase-api/model"
	"wx-purchase-api/repository"
)

const maxDescriptionLength = 50

type PurchaseTransactionUsecase struct {
	purchaseTransactionRepository repository.PurchaseTransactionRepository
}

func NewPurchaseTransactionUseCase(purchaseTransactionRepository repository.PurchaseTransactionRepository) PurchaseTransactionUsecase {
	return PurchaseTransactionUsecase{
		purchaseTransactionRepository,
	}
}

func (ru *PurchaseTransactionUsecase) GetTransactions() ([]model.PurchaseTransaction, error) {
	return ru.purchaseTransactionRepository.Get()
}

func (ru *PurchaseTransactionUsecase) SaveTransaction(transaction model.PurchaseTransaction) error {
	if err := validateTransaction(transaction); err != nil {
		return err
	}

	return ru.purchaseTransactionRepository.Save(transaction)
}

func validateTransaction(transaction model.PurchaseTransaction) error {
	if len(transaction.Description) > maxDescriptionLength {
		return fmt.Errorf("description must not be longer than %d characters", maxDescriptionLength)
	}

	if transaction.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	if _, err := time.Parse("01/02/2006", transaction.TransactionDate); err != nil {
		return fmt.Errorf("transactionDate must be in US format MM/DD/YYYY")
	}

	return nil
}

func (ru *PurchaseTransactionUsecase) GetTransactionById(id int) (model.PurchaseTransaction, error) {
	return ru.purchaseTransactionRepository.GetById(id)
}
