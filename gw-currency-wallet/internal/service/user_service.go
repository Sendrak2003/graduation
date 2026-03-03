package service

import (
	"context"
	"gw-currency-wallet/internal/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, userID string, username string, email string, password_hash string) error
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByID(ctx context.Context, userID string) (*models.User, error)
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(
	ctx context.Context,
	userID string,
	username string,
	email string,
	password_hash string,
) error {
	return s.repo.CreateUser(ctx, userID, username, email, password_hash)
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return s.repo.GetByUsername(ctx, username)
}

func (s *UserService) GetByID(ctx context.Context, userID string) (*models.User, error) {
	return s.repo.GetByID(ctx, userID)
}
