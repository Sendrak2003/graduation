package service

import (
	"gw-currency-wallet/internal/repository"
	"gw-currency-wallet/internal/utils/auth"

	"go.uber.org/zap"
)

type Services struct {
	UserService   *UserService
	WalletService *WalletService
}

func NewServices(repos *repository.Repositories, jwtManager *auth.Manager, logger *zap.Logger) *Services {
	return &Services{
		UserService:   NewUserService(repos.UserRepository, logger),
		WalletService: NewWalletService(repos.WalletRepository, logger),
	}
}
