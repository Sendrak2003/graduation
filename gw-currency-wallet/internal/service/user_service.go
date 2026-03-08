package service

import (
	"context"
	"gw-currency-wallet/pkg/models"

	"go.uber.org/zap"
)

type UserRepository interface {
	CreateUser(ctx context.Context, userID, username, email, passwordHash string) error
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByID(ctx context.Context, userID string) (*models.User, error)
}

type UserService struct {
	repo   UserRepository
	logger *zap.Logger
}

func NewUserService(repo UserRepository, logger *zap.Logger) *UserService {
	return &UserService{
		repo:   repo,
		logger: logger,
	}
}

func (s *UserService) Create(ctx context.Context, userID, username, email, passwordHash string) error {
	return s.repo.CreateUser(ctx, userID, username, email, passwordHash)
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return s.repo.GetByUsername(ctx, username)
}
