package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gw-notification/internal/kafka"
	"gw-notification/internal/logging"
	"gw-notification/internal/repository"
	"gw-notification/internal/service"

	"go.uber.org/zap"
)

func main() {
	logger, err := logging.NewLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Notification service starting...")

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:admin@localhost:27017"
	}

	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}

	logger.Info("MongoDB URI", zap.String("uri", mongoURI))
	logger.Info("Kafka Brokers", zap.String("brokers", kafkaBrokers))

	repo, err := repository.NewMongoRepository(mongoURI, "notifications")
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	logger.Info("MongoDB connected successfully")

	notificationService := service.NewNotificationService(repo)

	handleTransaction := func(ctx context.Context, tx *kafka.LargeTransaction) error {
		serviceTx := &service.Transaction{
			UserID:        tx.UserID,
			TransactionID: tx.TransactionID,
			Amount:        tx.Amount,
			Currency:      tx.Currency,
			Type:          tx.Type,
		}
		return notificationService.ProcessTransaction(ctx, serviceTx)
	}

	consumer, err := kafka.NewConsumer(
		kafkaBrokers,
		"notification-service",
		"large-transactions",
		handleTransaction,
		logger,
	)
	if err != nil {
		logger.Fatal("Failed to create Kafka consumer", zap.Error(err))
	}
	defer consumer.Close()
	logger.Info("Kafka consumer initialized")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := consumer.Start(ctx); err != nil && err != context.Canceled {
			logger.Error("Kafka consumer error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info("Shutting down notification service...")
	cancel()
	fmt.Println("Service exited gracefully")
}
