package service

import (
	"context"
	"errors"
	"gw-currency-wallet/internal/grpc"
	"gw-currency-wallet/internal/kafka"
	"gw-currency-wallet/internal/repository"
	"time"

	"go.uber.org/zap"
)

var ErrExchangeServiceUnavailable = errors.New("exchange service unavailable")

type ExchangeService struct {
	repo          *repository.WalletRepository
	grpcClient    *grpc.ExchangeClient
	kafkaProducer *kafka.Producer
	logger        *zap.Logger
}

func NewExchangeService(repo *repository.WalletRepository, grpcClient *grpc.ExchangeClient, kafkaProducer *kafka.Producer, logger *zap.Logger) *ExchangeService {
	return &ExchangeService{
		repo:          repo,
		grpcClient:    grpcClient,
		kafkaProducer: kafkaProducer,
		logger:        logger,
	}
}

func (s *ExchangeService) GetRates(ctx context.Context) (map[string]float64, error) {
	if s.grpcClient == nil {
		return nil, ErrExchangeServiceUnavailable
	}
	return s.grpcClient.GetAllExchangeRates(ctx)
}

func (s *ExchangeService) Exchange(ctx context.Context, userID, fromCurrency, toCurrency string, amount float64) (float64, error) {
	if s.grpcClient == nil {
		return 0, ErrExchangeServiceUnavailable
	}

	rates, err := s.grpcClient.GetAllExchangeRates(ctx)
	if err != nil {
		s.logger.Error("failed to get exchange rates", zap.Error(err))
		return 0, err
	}

	rateKey := fromCurrency + "_" + toCurrency
	rate, ok := rates[rateKey]
	if !ok {
		s.logger.Error("exchange rate not found", zap.String("from", fromCurrency), zap.String("to", toCurrency))
		return 0, repository.ErrWalletNotFound
	}

	toAmount := amount * rate

	if err := s.repo.ExchangeCurrency(ctx, userID, fromCurrency, toCurrency, amount, toAmount); err != nil {
		s.logger.Error("failed to exchange currency", zap.Error(err))
		return 0, err
	}

	if amount >= 1000 {
		msg := &kafka.LargeTransactionMessage{
			UserID:          userID,
			TransactionID:   userID + "_" + fromCurrency + "_" + toCurrency,
			FromCurrency:    fromCurrency,
			ToCurrency:      toCurrency,
			Amount:          amount,
			Currency:        fromCurrency,
			ExchangedAmount: toAmount,
			Rate:            rate,
			Type:            "exchange",
			Timestamp:       time.Now(),
		}
		s.kafkaProducer.SendTransaction(ctx, msg)
	}

	return toAmount, nil
}
