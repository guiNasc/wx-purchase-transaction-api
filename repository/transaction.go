package repository

import (
	"context"
	"database/sql"
	"fmt"
	"wx-purchase-api/model"
)

type PurchaseTransactionRepository struct {
	connection *sql.DB
}

func NewPurchaseTransactionRepository(connection *sql.DB) *PurchaseTransactionRepository {
	return &PurchaseTransactionRepository{
		connection,
	}
}

func (pr *PurchaseTransactionRepository) Get(ctx context.Context) ([]model.PurchaseTransaction, error) {

	qs := "SELECT id, description, amount, reference_date FROM purchase_transactions ORDER BY reference_date DESC	"
	rows, err := pr.connection.QueryContext(ctx, qs)
	if err != nil {
		fmt.Println(err)
		return []model.PurchaseTransaction{}, err
	}

	var transactionList []model.PurchaseTransaction
	var transactionObj model.PurchaseTransaction

	for rows.Next() {
		err = rows.Scan(
			&transactionObj.ID,
			&transactionObj.Description,
			&transactionObj.Amount,
			&transactionObj.TransactionDate,
		)

		if err != nil {
			fmt.Println(err)
			return []model.PurchaseTransaction{}, err
		}

		transactionList = append(transactionList, transactionObj)
	}

	rows.Close()

	return transactionList, nil

}

func (pr *PurchaseTransactionRepository) Save(ctx context.Context, transaction model.PurchaseTransaction) error {
	qs := "INSERT INTO purchase_transactions (description, amount, reference_date) VALUES ($1, $2, $3)"

	_, err := pr.connection.ExecContext(ctx, qs, transaction.Description, transaction.Amount, transaction.TransactionDate)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (pr *PurchaseTransactionRepository) GetById(ctx context.Context, id int) (model.PurchaseTransaction, error) {

	qs := "SELECT id, description, amount, reference_date FROM purchase_transactions WHERE id = $1"
	row := pr.connection.QueryRowContext(ctx, qs, id)

	var transactionObj model.PurchaseTransaction

	err := row.Scan(
		&transactionObj.ID,
		&transactionObj.Description,
		&transactionObj.Amount,
		&transactionObj.TransactionDate,
	)

	if err != nil {
		fmt.Println(err)
		return model.PurchaseTransaction{}, err
	}

	return transactionObj, nil

}
