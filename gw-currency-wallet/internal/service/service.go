package service

import (
	"gw-currency-wallet/internal/grpc"
	"gw-currency-wallet/internal/kafka"
	"gw-currency-wallet/internal/repository"
	"gw-currency-wallet/internal/utils/auth"

	"go.uber.org/zap"
)

type Services struct {
	UserService     *UserService
	WalletService   *WalletService
	ExchangeService *ExchangeService
}

func NewServices(repos *repository.Repositories, jwtManager *auth.Manager, kafkaProducer *kafka.Producer, grpcClient *grpc.ExchangeClient, logger *zap.Logger) *Services {
	return &Services{
		UserService:     NewUserService(repos.UserRepository, logger),
		WalletService:   NewWalletService(repos.WalletRepository, logger),
		ExchangeService: NewExchangeService(repos.WalletRepository, grpcClient, kafkaProducer, logger),
	}
}
