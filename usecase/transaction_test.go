package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"wx-purchase-api/model"
)

type purchaseRepositoryMock struct {
	getByIdFn func(ctx context.Context, id int) (model.PurchaseTransaction, error)
}

func (m *purchaseRepositoryMock) Get(ctx context.Context) ([]model.PurchaseTransaction, error) {
	return nil, nil
}

func (m *purchaseRepositoryMock) Save(ctx context.Context, transaction model.PurchaseTransaction) (model.PurchaseTransaction, error) {
	return transaction, nil
}

func (m *purchaseRepositoryMock) GetById(ctx context.Context, id int) (model.PurchaseTransaction, error) {
	if m.getByIdFn == nil {
		return model.PurchaseTransaction{}, nil
	}

	return m.getByIdFn(ctx, id)
}

type requestGatewayMock struct {
	getExchangeRateFn func(ctx context.Context, currency, fromDate, toDate string) (model.ExchangeAPIResponse, error)
}

func (m *requestGatewayMock) GetExchangeRate(ctx context.Context, currency, fromDate, toDate string) (model.ExchangeAPIResponse, error) {
	if m.getExchangeRateFn == nil {
		return model.ExchangeAPIResponse{}, nil
	}

	return m.getExchangeRateFn(ctx, currency, fromDate, toDate)
}

func TestValidateTransactionDescriptionLength(t *testing.T) {
	tests := []struct {
		name        string
		description string
		wantErr     bool
	}{
		{
			name:        "accepts description with exactly 50 chars",
			description: strings.Repeat("a", 50),
			wantErr:     false,
		},
		{
			name:        "rejects description with more than 50 chars",
			description: strings.Repeat("a", 51),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTransaction(model.PurchaseTransaction{
				Description:     tt.description,
				Amount:          10.0,
				TransactionDate: "2019-12-01",
			})
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got err: %v", tt.wantErr, err)
			}
		})
	}
}

func TestSaveTransactionReturnsValidationErrorWhenDescriptionIsTooLong(t *testing.T) {
	uc := &PurchaseTransactionUsecase{}

	_, err := uc.SaveTransaction(context.Background(), model.PurchaseTransaction{
		Description: strings.Repeat("a", 51),
	})

	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestValidateTransactionAmountMustBePositive(t *testing.T) {
	tests := []struct {
		name    string
		amount  float64
		wantErr bool
	}{
		{
			name:    "rejects zero amount",
			amount:  0,
			wantErr: true,
		},
		{
			name:    "rejects negative amount",
			amount:  -10.5,
			wantErr: true,
		},
		{
			name:    "accepts positive amount",
			amount:  10.5,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTransaction(model.PurchaseTransaction{
				Description:     "valid description",
				Amount:          tt.amount,
				TransactionDate: "2019-12-01",
			})

			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got err: %v", tt.wantErr, err)
			}
		})
	}
}

func TestValidateTransactionDateMustBeFormatted(t *testing.T) {
	tests := []struct {
		name            string
		transactionDate string
		wantErr         bool
	}{
		{
			name:            "accepts valid date",
			transactionDate: "2019-12-01",
			wantErr:         false,
		},
		{
			name:            "rejects non YYYY-MM-DD date format",
			transactionDate: "12/01/2019",
			wantErr:         true,
		},
		{
			name:            "rejects invalid date",
			transactionDate: "2019-13-40",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTransaction(model.PurchaseTransaction{
				Description:     "valid description",
				Amount:          100,
				TransactionDate: tt.transactionDate,
			})

			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got err: %v", tt.wantErr, err)
			}
		})
	}
}

func TestGetTransactionExchangeReturnsRepositoryError(t *testing.T) {
	repoErr := errors.New("repository failure")

	uc := &PurchaseTransactionUsecase{
		purchaseTransactionRepository: &purchaseRepositoryMock{
			getByIdFn: func(ctx context.Context, id int) (model.PurchaseTransaction, error) {
				return model.PurchaseTransaction{}, repoErr
			},
		},
		requestGateway: &requestGatewayMock{},
	}

	_, err := uc.GetTransactionExchange(context.Background(), 1, "Brazil-Real")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), repoErr.Error()) {
		t.Fatalf("expected error to contain %q, got %q", repoErr.Error(), err.Error())
	}
}

func TestGetTransactionExchangeReturnsErrorWhenDateIsInvalid(t *testing.T) {
	uc := &PurchaseTransactionUsecase{
		purchaseTransactionRepository: &purchaseRepositoryMock{
			getByIdFn: func(ctx context.Context, id int) (model.PurchaseTransaction, error) {
				return model.PurchaseTransaction{
					ID:              1,
					Description:     "test",
					Amount:          10,
					TransactionDate: "12/01/2019",
				}, nil
			},
		},
		requestGateway: &requestGatewayMock{},
	}

	_, err := uc.GetTransactionExchange(context.Background(), 1, "Brazil-Real")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "invalid date format") {
		t.Fatalf("expected invalid date error, got %q", err.Error())
	}
}

func TestGetTransactionExchangeReturnsGatewayError(t *testing.T) {
	gatewayErr := errors.New("gateway failure")

	uc := &PurchaseTransactionUsecase{
		purchaseTransactionRepository: &purchaseRepositoryMock{
			getByIdFn: func(ctx context.Context, id int) (model.PurchaseTransaction, error) {
				return model.PurchaseTransaction{
					ID:              1,
					Description:     "test",
					Amount:          10,
					TransactionDate: "2019-07-25",
				}, nil
			},
		},
		requestGateway: &requestGatewayMock{
			getExchangeRateFn: func(ctx context.Context, currency, fromDate, toDate string) (model.ExchangeAPIResponse, error) {
				return model.ExchangeAPIResponse{}, gatewayErr
			},
		},
	}

	_, err := uc.GetTransactionExchange(context.Background(), 1, "Brazil-Real")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), gatewayErr.Error()) {
		t.Fatalf("expected error to contain %q, got %q", gatewayErr.Error(), err.Error())
	}
}

func TestGetTransactionExchangeCallsGatewayWithExpectedParams(t *testing.T) {
	called := false

	uc := &PurchaseTransactionUsecase{
		purchaseTransactionRepository: &purchaseRepositoryMock{
			getByIdFn: func(ctx context.Context, id int) (model.PurchaseTransaction, error) {
				return model.PurchaseTransaction{
					ID:              1,
					Description:     "test",
					Amount:          10,
					TransactionDate: "2019-07-25",
				}, nil
			},
		},
		requestGateway: &requestGatewayMock{
			getExchangeRateFn: func(ctx context.Context, currency, fromDate, toDate string) (model.ExchangeAPIResponse, error) {
				called = true

				if currency != "Brazil-Real" {
					t.Fatalf("expected currency Brazil-Real, got %s", currency)
				}

				if fromDate != "2019-01-25" {
					t.Fatalf("expected from date 2019-01-25, got %s", fromDate)
				}

				if toDate != "2019-07-25" {
					t.Fatalf("expected to date 2019-07-25, got %s", toDate)
				}

				reponse := model.ExchangeAPIResponse{
					Meta: model.ExchangeAPIResponseMeta{
						Count: 1,
					},
					Data: []model.ExchangeAPIObj{
						{
							Currency:     "Real",
							ExchangeRate: "5.34",
						},
					},
				}

				return reponse, nil
			},
		},
	}

	_, err := uc.GetTransactionExchange(context.Background(), 1, "Brazil-Real")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !called {
		t.Fatal("expected gateway to be called")
	}
}

func TestGetTransactionExchangeNoConversionAvailable(t *testing.T) {
	called := false
	wantErr := errors.New("no exchange rate data found for the requested period")

	uc := &PurchaseTransactionUsecase{
		purchaseTransactionRepository: &purchaseRepositoryMock{
			getByIdFn: func(ctx context.Context, id int) (model.PurchaseTransaction, error) {
				return model.PurchaseTransaction{
					ID:              1,
					Description:     "test",
					Amount:          10,
					TransactionDate: "2019-07-25",
				}, nil
			},
		},
		requestGateway: &requestGatewayMock{
			getExchangeRateFn: func(ctx context.Context, currency, fromDate, toDate string) (model.ExchangeAPIResponse, error) {
				called = true

				reponse := model.ExchangeAPIResponse{
					Meta: model.ExchangeAPIResponseMeta{
						Count: 0,
					},
				}

				return reponse, nil
			},
		},
	}

	_, err := uc.GetTransactionExchange(context.Background(), 1, "Brazil-Real")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !strings.Contains(err.Error(), wantErr.Error()) {
		t.Fatalf("expected error to contain %q, got %q", wantErr.Error(), err.Error())
	}

	if !called {
		t.Fatal("expected gateway to be called")
	}
}
