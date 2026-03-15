package usecase

import (
	"wx-purchase-api/model"
	"wx-purchase-api/repository"
)

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
	return ru.purchaseTransactionRepository.Save(transaction)
}
