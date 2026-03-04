package service

import (
	"context"
	"gw-currency-wallet/internal/repository"

	"go.uber.org/zap"
)

type WalletService struct {
	repo   *repository.WalletRepository
	logger *zap.Logger
}

func NewWalletService(repo *repository.WalletRepository, logger *zap.Logger) *WalletService {
	return &WalletService{
		repo:   repo,
		logger: logger,
	}
}

func (s *WalletService) GetBalance(ctx context.Context, userID string, currency string) (float64, error) {
	s.logger.Debug("getting balance",
		zap.String("user_id", userID),
		zap.String("currency", currency))
	return s.repo.GetBalance(ctx, userID, currency)
}

func (s *WalletService) GetAllBalances(ctx context.Context, userID string) (map[string]float64, error) {
	s.logger.Debug("getting all balances", zap.String("user_id", userID))

	currencies := []string{"USD", "EUR", "RUB"}
	balances := make(map[string]float64)

	for _, currency := range currencies {
		balance, err := s.repo.GetBalance(ctx, userID, currency)
		if err != nil {
			s.logger.Error("failed to get balance for currency",
				zap.String("user_id", userID),
				zap.String("currency", currency),
				zap.Error(err))
			return nil, err
		}
		balances[currency] = balance
	}

	return balances, nil
}

func (s *WalletService) Deposit(ctx context.Context, userID string, currency string, amount float64) error {
	s.logger.Debug("depositing funds",
		zap.String("user_id", userID),
		zap.String("currency", currency),
		zap.Float64("amount", amount))
	return s.repo.Deposit(ctx, userID, currency, amount)
}

func (s *WalletService) Withdraw(ctx context.Context, userID string, currency string, amount float64) error {
	s.logger.Debug("withdrawing funds",
		zap.String("user_id", userID),
		zap.String("currency", currency),
		zap.Float64("amount", amount))
	return s.repo.Withdraw(ctx, userID, currency, amount)
}
