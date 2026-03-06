package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"gw-notification/internal/kafka"
	"gw-notification/internal/repository"
	"gw-notification/internal/service"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("notification service starting")

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:admin@mongodb:27017"
	}

	mongoDB := os.Getenv("MONGO_DB")
	if mongoDB == "" {
		mongoDB = "notifications"
	}

	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "kafka:9092"
	}

	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "large-transactions"
	}

	kafkaGroupID := os.Getenv("KAFKA_GROUP_ID")
	if kafkaGroupID == "" {
		kafkaGroupID = "notification-service"
	}

	mongoRepo, err := repository.NewMongoRepository(mongoURI, mongoDB, logger)
	if err != nil {
		logger.Fatal("failed to create mongo repository", zap.Error(err))
	}

	notificationService := service.NewNotificationService(mongoRepo, logger)

	handler := func(ctx context.Context, tx *kafka.LargeTransaction) error {
		doc := &repository.LargeTransactionDoc{
			UserID:          tx.UserID,
			TransactionID:   tx.TransactionID,
			Amount:          tx.Amount,
			Currency:        tx.Currency,
			FromCurrency:    tx.FromCurrency,
			ToCurrency:      tx.ToCurrency,
			ExchangedAmount: tx.ExchangedAmount,
			Rate:            tx.Rate,
			Type:            tx.Type,
			Timestamp:       tx.Timestamp,
		}
		return notificationService.ProcessTransaction(ctx, doc)
	}

	consumer, err := kafka.NewConsumer(kafkaBrokers, kafkaGroupID, kafkaTopic, handler, logger)
	if err != nil {
		logger.Fatal("failed to create kafka consumer", zap.Error(err))
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := consumer.Start(ctx); err != nil {
			logger.Error("kafka consumer error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down notification service")
	cancel()
	logger.Info("service exited gracefully")
}
