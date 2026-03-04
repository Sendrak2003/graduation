package service

import (
	"context"
	"gw-currency-wallet/pkg/models"

	"go.uber.org/zap"
)

type UserRepository interface {
	CreateUser(ctx context.Context, userID string, username string, email string, password_hash string) error
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

func (s *UserService) Create(
	ctx context.Context,
	userID string,
	username string,
	email string,
	password_hash string,
) error {
	s.logger.Debug("creating user",
		zap.String("user_id", userID),
		zap.String("username", username),
		zap.String("email", email))
	return s.repo.CreateUser(ctx, userID, username, email, password_hash)
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	s.logger.Debug("getting user by username", zap.String("username", username))
	return s.repo.GetByUsername(ctx, username)
}

func (s *UserService) GetByID(ctx context.Context, userID string) (*models.User, error) {
	s.logger.Debug("getting user by ID", zap.String("user_id", userID))
	return s.repo.GetByID(ctx, userID)
}
