package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"wx-purchase-api/apperror"
	"wx-purchase-api/model"

	"github.com/lib/pq"
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

	qs := "SELECT id, description, amount, TO_CHAR(reference_date, 'YYYY-MM-DD') FROM purchase_transactions ORDER BY reference_date DESC	"
	rows, err := pr.connection.QueryContext(ctx, qs)
	if err != nil {
		slog.Error("failed to query transactions", "error", err)
		return []model.PurchaseTransaction{}, err
	}
	defer rows.Close()

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
			slog.Error("failed to scan transaction row", "error", err)
			return []model.PurchaseTransaction{}, err
		}

		transactionList = append(transactionList, transactionObj)
	}

	if err := rows.Err(); err != nil {
		slog.Error("transaction rows iteration failed", "error", err)
		return []model.PurchaseTransaction{}, err
	}

	return transactionList, nil

}

func (pr *PurchaseTransactionRepository) Save(ctx context.Context, transaction model.PurchaseTransaction) (model.PurchaseTransaction, error) {
	qs := "INSERT INTO purchase_transactions (description, amount, reference_date) VALUES ($1, $2, $3) RETURNING id"

	err := pr.connection.QueryRowContext(ctx, qs, transaction.Description,
		transaction.Amount, transaction.TransactionDate,
	).Scan(&transaction.ID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return model.PurchaseTransaction{}, apperror.Conflict("transaction_conflict", "transaction conflicts with an existing record", err)
		}

		slog.Error("failed to insert transaction", "error", err)
		return model.PurchaseTransaction{}, err
	}

	return transaction, nil
}

func (pr *PurchaseTransactionRepository) GetById(ctx context.Context, id int) (model.PurchaseTransaction, error) {

	qs := "SELECT id, description, amount, TO_CHAR(reference_date, 'YYYY-MM-DD') FROM purchase_transactions WHERE id = $1"
	row := pr.connection.QueryRowContext(ctx, qs, id)

	var transactionObj model.PurchaseTransaction

	err := row.Scan(
		&transactionObj.ID,
		&transactionObj.Description,
		&transactionObj.Amount,
		&transactionObj.TransactionDate,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PurchaseTransaction{}, apperror.NotFound("transaction_not_found", "transaction not found", err)
		}

		slog.Error("failed to get transaction by id", "error", err, "id", id)
		return model.PurchaseTransaction{}, err
	}

	return transactionObj, nil

}
