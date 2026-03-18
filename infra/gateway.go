package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"wx-purchase-api/model"
)

type Gateway struct {
	client *http.Client
}

func NewGateway() *Gateway {
	return &Gateway{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (g *Gateway) GetExchangeRate(ctx context.Context, currency, fromDate, toDate string) (model.ExchangeAPIResponse, error) {

	baseUrl := "https://api.fiscaldata.treasury.gov/services/api/fiscal_service/"
	endpoint := fmt.Sprintf("%sv1/accounting/od/rates_of_exchange", baseUrl)
	url := fmt.Sprintf("%s?fields=record_date,country,currency,country_currency_desc,exchange_rate", endpoint)
	url += fmt.Sprintf("&filter=currency:eq:%s,record_date:gte:%s,record_date:lte:%s&sort=-record_date", currency, fromDate,
		toDate,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	var payload model.ExchangeAPIResponse

	if err != nil {
		slog.Error("failed to create exchange API request", "error", err, "currency", currency, "from_date", fromDate, "to_date", toDate)
		return payload, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "wx-purchase-api/1.0")

	resp, err := g.client.Do(req)
	if err != nil {
		slog.Error("failed to call exchange API", "error", err, "currency", currency, "from_date", fromDate, "to_date", toDate)
		return payload, fmt.Errorf("failed to call exchange endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("exchange API returned non-200 status", "status", resp.StatusCode, "currency", currency)
		return payload, fmt.Errorf("exchange endpoint returned status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		slog.Error("failed to decode exchange API response", "error", err, "currency", currency)
		return payload, fmt.Errorf("failed to decode exchange response: %w", err)
	}

	slog.Debug("exchange API response decoded", "currency", currency, "records", payload.Meta.Count)

	return payload, nil
}
