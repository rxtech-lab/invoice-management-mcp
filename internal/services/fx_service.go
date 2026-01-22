package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	frankfurterBaseURL = "https://api.frankfurter.dev/v1"
	defaultCacheTTL    = time.Minute
)

// ExchangeRate represents an exchange rate between two currencies
type ExchangeRate struct {
	From     string    `json:"from"`
	To       string    `json:"to"`
	Rate     float64   `json:"rate"`
	Date     string    `json:"date"`
	CachedAt time.Time `json:"cached_at"`
}

// FXService provides currency conversion functionality
type FXService interface {
	// GetExchangeRate gets the exchange rate from one currency to another
	GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (*ExchangeRate, error)
	// ConvertAmount converts an amount from one currency to another
	// Returns (convertedAmount, rateUsed, error)
	ConvertAmount(ctx context.Context, amount float64, fromCurrency, toCurrency string) (float64, float64, error)
}

type fxService struct {
	redis      RedisService
	httpClient *http.Client
	cacheTTL   time.Duration
}

// frankfurterResponse represents the API response from Frankfurter
type frankfurterResponse struct {
	Amount float64            `json:"amount"`
	Base   string             `json:"base"`
	Date   string             `json:"date"`
	Rates  map[string]float64 `json:"rates"`
}

// NewFXService creates a new FX service
// redis can be nil (caching will be disabled)
func NewFXService(redis RedisService, httpClient *http.Client, cacheTTL time.Duration) FXService {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	if cacheTTL == 0 {
		cacheTTL = defaultCacheTTL
	}
	return &fxService{
		redis:      redis,
		httpClient: httpClient,
		cacheTTL:   cacheTTL,
	}
}

func (f *fxService) cacheKey(from, to string) string {
	return fmt.Sprintf("fx_rate:%s:%s", strings.ToUpper(from), strings.ToUpper(to))
}

func (f *fxService) GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (*ExchangeRate, error) {
	from := strings.ToUpper(fromCurrency)
	to := strings.ToUpper(toCurrency)

	// Same currency - return 1.0
	if from == to {
		return &ExchangeRate{
			From:     from,
			To:       to,
			Rate:     1.0,
			Date:     time.Now().Format("2006-01-02"),
			CachedAt: time.Now(),
		}, nil
	}

	// Try to get from cache
	if f.redis != nil {
		cached, err := f.redis.Get(ctx, f.cacheKey(from, to))
		if err == nil && cached != "" {
			var rate ExchangeRate
			if err := json.Unmarshal([]byte(cached), &rate); err == nil {
				return &rate, nil
			}
		}
	}

	// Fetch from API
	rate, err := f.fetchFromAPI(ctx, from, to)
	if err != nil {
		log.Printf("Warning: Failed to fetch exchange rate from API: %v", err)
		// Return 1.0 rate on failure (fail-soft)
		return &ExchangeRate{
			From:     from,
			To:       to,
			Rate:     1.0,
			Date:     time.Now().Format("2006-01-02"),
			CachedAt: time.Now(),
		}, nil
	}

	// Cache the result
	if f.redis != nil {
		data, err := json.Marshal(rate)
		if err == nil {
			if err := f.redis.Set(ctx, f.cacheKey(from, to), string(data), f.cacheTTL); err != nil {
				log.Printf("Warning: Failed to cache exchange rate: %v", err)
			}
		}
	}

	return rate, nil
}

func (f *fxService) fetchFromAPI(ctx context.Context, from, to string) (*ExchangeRate, error) {
	url := fmt.Sprintf("%s/latest?base=%s&symbols=%s", frankfurterBaseURL, from, to)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exchange rate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var apiResp frankfurterResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	rate, ok := apiResp.Rates[to]
	if !ok {
		return nil, fmt.Errorf("rate for %s not found in response", to)
	}

	return &ExchangeRate{
		From:     from,
		To:       to,
		Rate:     rate,
		Date:     apiResp.Date,
		CachedAt: time.Now(),
	}, nil
}

func (f *fxService) ConvertAmount(ctx context.Context, amount float64, fromCurrency, toCurrency string) (float64, float64, error) {
	rate, err := f.GetExchangeRate(ctx, fromCurrency, toCurrency)
	if err != nil {
		// This shouldn't happen as GetExchangeRate always returns a valid rate
		// but handle it just in case
		return amount, 1.0, nil
	}

	convertedAmount := amount * rate.Rate
	return convertedAmount, rate.Rate, nil
}
