package service

import (
	"context"
	"errors"
	"testing"
)

type mockStorage struct {
	rates      map[string]float64
	allRates   map[string]float64
	updateErr  error
	getRateErr error
}

func (m *mockStorage) GetExchangeRate(ctx context.Context, from, to string) (float64, error) {
	if m.getRateErr != nil {
		return 0, m.getRateErr
	}
	key := from + "_" + to
	if rate, ok := m.rates[key]; ok {
		return rate, nil
	}
	return 0, errors.New("rate not found")
}

func (m *mockStorage) GetAllExchangeRates(ctx context.Context) (map[string]float64, error) {
	if m.allRates == nil {
		return make(map[string]float64), nil
	}
	return m.allRates, nil
}

func (m *mockStorage) UpdateExchangeRate(ctx context.Context, from, to string, rate float64) error {
	return m.updateErr
}

func TestExchangeService_GetExchangeRate(t *testing.T) {
	storage := &mockStorage{
		rates: map[string]float64{
			"USD_RUB": 90.5,
			"EUR_USD": 1.09,
		},
	}

	service := NewExchangeService(storage)
	ctx := context.Background()

	rate, err := service.GetExchangeRate(ctx, "USD", "RUB")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rate != 90.5 {
		t.Errorf("expected rate 90.5, got %f", rate)
	}
}

func TestExchangeService_GetExchangeRate_NotFound(t *testing.T) {
	storage := &mockStorage{
		rates: map[string]float64{},
	}

	service := NewExchangeService(storage)
	ctx := context.Background()

	_, err := service.GetExchangeRate(ctx, "USD", "RUB")
	if err == nil {
		t.Error("expected error for non-existent rate, got nil")
	}
}

func TestExchangeService_GetAllExchangeRates(t *testing.T) {
	storage := &mockStorage{
		allRates: map[string]float64{
			"USD_RUB": 90.5,
			"EUR_USD": 1.09,
			"USD_EUR": 0.92,
		},
	}

	service := NewExchangeService(storage)
	ctx := context.Background()

	rates, err := service.GetAllExchangeRates(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(rates) != 3 {
		t.Errorf("expected 3 rates, got %d", len(rates))
	}

	if rates["USD_RUB"] != 90.5 {
		t.Errorf("expected USD_RUB rate 90.5, got %f", rates["USD_RUB"])
	}
}

func TestExchangeService_GetAllExchangeRates_Empty(t *testing.T) {
	storage := &mockStorage{
		allRates: map[string]float64{},
	}

	service := NewExchangeService(storage)
	ctx := context.Background()

	rates, err := service.GetAllExchangeRates(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(rates) != 0 {
		t.Errorf("expected 0 rates, got %d", len(rates))
	}
}
