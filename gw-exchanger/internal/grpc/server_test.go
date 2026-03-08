package grpc

import (
	"context"
	"errors"
	"testing"

	pb "proto-exchange/exchange"

	"go.uber.org/zap"
)

type mockExchangeService struct {
	rate     float64
	allRates map[string]float64
	err      error
}

func (m *mockExchangeService) GetExchangeRate(ctx context.Context, from, to string) (float64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.rate, nil
}

func (m *mockExchangeService) GetAllExchangeRates(ctx context.Context) (map[string]float64, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.allRates, nil
}

func TestExchangeServer_GetExchangeRateForCurrency(t *testing.T) {
	logger := zap.NewNop()
	mockService := &mockExchangeService{
		rate: 90.5,
	}

	server := NewExchangeServer(mockService, logger)
	ctx := context.Background()

	req := &pb.CurrencyRequest{
		FromCurrency: "USD",
		ToCurrency:   "RUB",
	}

	resp, err := server.GetExchangeRateForCurrency(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.FromCurrency != "USD" {
		t.Errorf("expected FromCurrency USD, got %s", resp.FromCurrency)
	}

	if resp.ToCurrency != "RUB" {
		t.Errorf("expected ToCurrency RUB, got %s", resp.ToCurrency)
	}

	if resp.Rate != 90.5 {
		t.Errorf("expected rate 90.5, got %f", resp.Rate)
	}
}

func TestExchangeServer_GetExchangeRateForCurrency_Error(t *testing.T) {
	logger := zap.NewNop()
	mockService := &mockExchangeService{
		err: errors.New("rate not found"),
	}

	server := NewExchangeServer(mockService, logger)
	ctx := context.Background()

	req := &pb.CurrencyRequest{
		FromCurrency: "USD",
		ToCurrency:   "XXX",
	}

	_, err := server.GetExchangeRateForCurrency(ctx, req)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestExchangeServer_GetExchangeRates(t *testing.T) {
	logger := zap.NewNop()
	mockService := &mockExchangeService{
		allRates: map[string]float64{
			"USD_RUB": 90.5,
			"USD_EUR": 0.92,
			"EUR_RUB": 98.3,
		},
	}

	server := NewExchangeServer(mockService, logger)
	ctx := context.Background()

	resp, err := server.GetExchangeRates(ctx, &pb.Empty{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(resp.Rates) != 3 {
		t.Errorf("expected 3 rates, got %d", len(resp.Rates))
	}

	if resp.Rates["USD_RUB"] != 90.5 {
		t.Errorf("expected USD_RUB rate 90.5, got %f", resp.Rates["USD_RUB"])
	}
}

func TestExchangeServer_GetExchangeRates_Error(t *testing.T) {
	logger := zap.NewNop()
	mockService := &mockExchangeService{
		err: errors.New("database error"),
	}

	server := NewExchangeServer(mockService, logger)
	ctx := context.Background()

	_, err := server.GetExchangeRates(ctx, &pb.Empty{})
	if err == nil {
		t.Error("expected error, got nil")
	}
}
