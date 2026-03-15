package model

type PurchaseTransaction struct {
	Description     string  `json:"description"`
	TransactionDate string  `json:"transactionDate"`
	Amount          float64 `json:"amount"`
	ID              string  `json:"id"`
}

type HealthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}
