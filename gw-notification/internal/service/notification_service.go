package service

import (
	"context"
	"gw-notification/internal/repository"

	"go.uber.org/zap"
)

type NotificationService struct {
	repo   *repository.MongoRepository
	logger *zap.Logger
}

func NewNotificationService(repo *repository.MongoRepository, logger *zap.Logger) *NotificationService {
	return &NotificationService{
		repo:   repo,
		logger: logger,
	}
}

func (s *NotificationService) ProcessTransaction(ctx context.Context, tx *repository.LargeTransactionDoc) error {
	s.logger.Info("processing large transaction",
		zap.String("transaction_id", tx.TransactionID),
		zap.String("user_id", tx.UserID),
		zap.Float64("amount", tx.Amount),
		zap.String("from_currency", tx.FromCurrency),
		zap.String("to_currency", tx.ToCurrency))

	if err := s.repo.SaveTransaction(ctx, tx); err != nil {
		s.logger.Error("failed to save transaction",
			zap.String("transaction_id", tx.TransactionID),
			zap.Error(err))
		return err
	}

	s.logger.Info("transaction saved successfully",
		zap.String("transaction_id", tx.TransactionID))
	return nil
}
