package usecase

import (
	"strings"
	"testing"

	"wx-purchase-api/model"
)

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

	err := uc.SaveTransaction(model.PurchaseTransaction{
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

func TestValidateTransactionDateMustBeUSFormat(t *testing.T) {
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
