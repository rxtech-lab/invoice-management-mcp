package services

import (
	"context"
	"time"
)

// MockFXService implements FXService with configurable exchange rates for testing
type MockFXService struct {
	rates map[string]float64 // Key format: "FROM:TO"
}

// NewMockFXService creates a new mock FX service
func NewMockFXService() *MockFXService {
	return &MockFXService{
		rates: make(map[string]float64),
	}
}

// SetRate sets the exchange rate from one currency to another
func (m *MockFXService) SetRate(from, to string, rate float64) {
	key := from + ":" + to
	m.rates[key] = rate
}

// GetExchangeRate implements FXService
func (m *MockFXService) GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (*ExchangeRate, error) {
	if fromCurrency == toCurrency {
		return &ExchangeRate{
			From:     fromCurrency,
			To:       toCurrency,
			Rate:     1.0,
			Date:     time.Now().Format("2006-01-02"),
			CachedAt: time.Now(),
		}, nil
	}

	key := fromCurrency + ":" + toCurrency
	rate, ok := m.rates[key]
	if !ok {
		// Default to 1.0 if rate not configured (fail-soft like real service)
		rate = 1.0
	}

	return &ExchangeRate{
		From:     fromCurrency,
		To:       toCurrency,
		Rate:     rate,
		Date:     time.Now().Format("2006-01-02"),
		CachedAt: time.Now(),
	}, nil
}

// ConvertAmount implements FXService
func (m *MockFXService) ConvertAmount(ctx context.Context, amount float64, fromCurrency, toCurrency string) (float64, float64, error) {
	rate, err := m.GetExchangeRate(ctx, fromCurrency, toCurrency)
	if err != nil {
		return amount, 1.0, nil
	}
	return amount * rate.Rate, rate.Rate, nil
}
