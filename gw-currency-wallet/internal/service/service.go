package service

import (
	"gw-currency-wallet/internal/repository"
	"gw-currency-wallet/internal/utils/auth"
)

type Services struct {
	UserService   *UserService
	WalletService *WalletService
}

func NewServices(repos *repository.Repositories, jwtManager *auth.Manager) *Services {
	return &Services{
		UserService:   NewUserService(repos.UserRepository),
		WalletService: NewWalletService(repos.WalletRepository),
	}
}
