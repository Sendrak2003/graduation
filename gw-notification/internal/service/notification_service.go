package service

import (
	"context"
	"gw-notification/internal/repository"
	"log"
)

type Transaction struct {
	UserID        string
	TransactionID string
	Amount        float64
	Currency      string
	Type          string
	Timestamp     string
}

type NotificationService struct {
	repo *repository.MongoRepository
}

func NewNotificationService(repo *repository.MongoRepository) *NotificationService {
	return &NotificationService{
		repo: repo,
	}
}

func (s *NotificationService) ProcessTransaction(ctx context.Context, tx *Transaction) error {
	log.Printf("Processing transaction: %s, amount: %.2f %s", tx.TransactionID, tx.Amount, tx.Currency)

	// Convert to repository document
	doc := &repository.LargeTransactionDoc{
		UserID:        tx.UserID,
		TransactionID: tx.TransactionID,
		Amount:        tx.Amount,
		Currency:      tx.Currency,
		Type:          tx.Type,
	}

	if err := s.repo.SaveTransaction(ctx, doc); err != nil {
		return err
	}

	log.Printf("Transaction %s saved successfully", tx.TransactionID)
	return nil
}
