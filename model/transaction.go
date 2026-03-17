package model

type PurchaseTransaction struct {
	Description     string  `json:"description"`
	TransactionDate string  `json:"transactionDate"`
	Amount          float64 `json:"amount"`
	ID              string  `json:"id"`
}

type PurchaseTransactionExchange struct {
	Description     string  `json:"description"`
	TransactionDate string  `json:"transactionDate"`
	Amount          float64 `json:"amount"`
	ID              string  `json:"id"`
	Currency        string  `json:"currency"`
	ExchangeRate    float64 `json:"exchangeRate"`
	ConvertedAmount float64 `json:"convertedAmount"`
}

type ExchangeAPIObj struct {
	RecordDate          string `json:"record_date"`
	Country             string `json:"country"`
	Currency            string `json:"currency"`
	CountryCurrencyDesc string `json:"country_currency_desc"`
	ExchangeRate        string `json:"exchange_rate"`
	EffectiveDate       string `json:"effective_date"`
}

type ExchangeAPIResponse struct {
	Data []ExchangeAPIObj `json:"data"`
}

type HealthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}
