package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"wx-purchase-api/model"
)

type Gateway struct {
	client *http.Client
}

func NewGateway() *Gateway {
	return &Gateway{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (g *Gateway) GetExchangeRate(ctx context.Context, endpoint string, currency string) (model.ExchangeAPIResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	var payload model.ExchangeAPIResponse

	if err != nil {
		return payload, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return payload, fmt.Errorf("failed to call exchange endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return payload, fmt.Errorf("exchange endpoint returned status %d", resp.StatusCode)
	}

	fmt.Println("Received response from exchange API")
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return payload, fmt.Errorf("failed to decode exchange response: %w", err)
	}

	//TODO remove this later, just to check the response from the API
	// rate, ok := payload.Data[currency]
	// if !ok {
	// 	return pte, fmt.Errorf("currency %s not found in response", currency)
	// }

	return payload, nil
}
