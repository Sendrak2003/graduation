package storages

import "context"

type Storage interface {
	GetExchangeRate(ctx context.Context, from, to string) (float64, error)
	GetAllExchangeRates(ctx context.Context) (map[string]float64, error)
	UpdateExchangeRate(ctx context.Context, from, to string, rate float64) error
}
