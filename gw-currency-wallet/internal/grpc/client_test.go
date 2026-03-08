package grpc

import (
	"context"
	"testing"

	pb "proto-exchange/exchange"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type mockExchangeServiceClient struct {
	rates map[string]float32
	err   error
}

func (m *mockExchangeServiceClient) GetExchangeRates(ctx context.Context, in *pb.Empty, opts ...grpc.CallOption) (*pb.ExchangeRatesResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &pb.ExchangeRatesResponse{
		Rates: m.rates,
	}, nil
}

func (m *mockExchangeServiceClient) GetExchangeRateForCurrency(ctx context.Context, in *pb.CurrencyRequest, opts ...grpc.CallOption) (*pb.ExchangeRateResponse, error) {
	return nil, nil
}

func TestExchangeClient_GetAllExchangeRates(t *testing.T) {
	logger := zap.NewNop()
	mockClient := &mockExchangeServiceClient{
		rates: map[string]float32{
			"USD_RUB": 90.5,
			"USD_EUR": 0.92,
			"EUR_RUB": 98.3,
		},
	}

	client := &ExchangeClient{
		client: mockClient,
		logger: logger,
	}

	ctx := context.Background()
	rates, err := client.GetAllExchangeRates(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(rates) != 3 {
		t.Errorf("expected 3 rates, got %d", len(rates))
	}

	if rates["USD_RUB"] != 90.5 {
		t.Errorf("expected USD_RUB rate 90.5, got %f", rates["USD_RUB"])
	}

	if rates["USD_EUR"] != 0.92 {
		t.Errorf("expected USD_EUR rate 0.92, got %f", rates["USD_EUR"])
	}
}

func TestExchangeClient_GetAllExchangeRates_Empty(t *testing.T) {
	logger := zap.NewNop()
	mockClient := &mockExchangeServiceClient{
		rates: map[string]float32{},
	}

	client := &ExchangeClient{
		client: mockClient,
		logger: logger,
	}

	ctx := context.Background()
	rates, err := client.GetAllExchangeRates(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(rates) != 0 {
		t.Errorf("expected 0 rates, got %d", len(rates))
	}
}
