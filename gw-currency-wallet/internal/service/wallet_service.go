package service

import (
	"context"
	"gw-currency-wallet/internal/repository"
)

type WalletService struct {
	repo *repository.WalletRepository
}

func NewWalletService(repo *repository.WalletRepository) *WalletService {
	return &WalletService{repo: repo}
}

func (s *WalletService) GetBalance(ctx context.Context, userID string, currency string) (float64, error) {
	return s.repo.GetBalance(ctx, userID, currency)
}

func (s *WalletService) Deposit(ctx context.Context, userID string, currency string, amount float64) error {
	return s.repo.Deposit(ctx, userID, currency, amount)
}

func (s *WalletService) Withdraw(ctx context.Context, userID string, currency string, amount float64) error {
	return s.repo.Withdraw(ctx, userID, currency, amount)
}
