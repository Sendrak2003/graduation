package service

import (
	"context"
	"gw-exchanger/internal/storages"
)

type ExchangeServiceInterface interface {
	GetExchangeRate(ctx context.Context, from, to string) (float64, error)
	GetAllExchangeRates(ctx context.Context) (map[string]float64, error)
}

type ExchangeService struct {
	storage storages.Storage
}

func NewExchangeService(storage storages.Storage) *ExchangeService {
	return &ExchangeService{
		storage: storage,
	}
}

func (s *ExchangeService) GetExchangeRate(ctx context.Context, from, to string) (float64, error) {
	return s.storage.GetExchangeRate(ctx, from, to)
}

func (s *ExchangeService) GetAllExchangeRates(ctx context.Context) (map[string]float64, error) {
	return s.storage.GetAllExchangeRates(ctx)
}
