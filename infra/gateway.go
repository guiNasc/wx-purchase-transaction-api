package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"
	"wx-purchase-api/apperror"
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

	baseURL := "https://api.fiscaldata.treasury.gov/services/api/fiscal_service/"
	endpoint := fmt.Sprintf("%sv1/accounting/od/rates_of_exchange", baseURL)
	requestURL := fmt.Sprintf("%s?fields=record_date,country,currency,country_currency_desc,exchange_rate", endpoint)
	requestURL += fmt.Sprintf("&filter=currency:eq:%s,record_date:gte:%s,record_date:lte:%s&sort=-record_date", url.QueryEscape(currency), fromDate,
		toDate,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	var payload model.ExchangeAPIResponse

	if err != nil {
		slog.Error("failed to create exchange API request", "error", err, "currency", currency, "from_date", fromDate, "to_date", toDate)
		return payload, apperror.ServiceUnavailable("exchange_request_build_failed", "failed to build exchange rate request", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "wx-purchase-api/1.0")

	resp, err := g.client.Do(req)
	if err != nil {
		slog.Error("failed to call exchange API", "error", err, "currency", currency, "from_date", fromDate, "to_date", toDate)
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return payload, apperror.ServiceUnavailable("exchange_timeout", "exchange service timed out", err)
		}

		return payload, apperror.ServiceUnavailable("exchange_unavailable", "exchange service is unavailable", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("exchange API returned non-200 status", "status", resp.StatusCode, "currency", currency)
		switch resp.StatusCode {
		case http.StatusTooManyRequests:
			return payload, apperror.RateLimited("exchange_rate_limited", "exchange service rate limit exceeded", nil)
		case http.StatusBadRequest, http.StatusNotFound, http.StatusUnprocessableEntity:
			return payload, apperror.Unprocessable("invalid_currency", "currency is invalid or not supported", nil)
		case http.StatusServiceUnavailable, http.StatusGatewayTimeout, http.StatusBadGateway:
			return payload, apperror.ServiceUnavailable("exchange_unavailable", "exchange service is unavailable", nil)
		default:
			return payload, apperror.ServiceUnavailable("exchange_unavailable", "exchange service returned an unexpected status", nil)
		}
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		slog.Error("failed to decode exchange API response", "error", err, "currency", currency)
		return payload, apperror.ServiceUnavailable("invalid_exchange_response", "failed to decode exchange service response", err)
	}

	slog.Debug("exchange API response decoded", "currency", currency, "records", payload.Meta.Count)

	return payload, nil
}
