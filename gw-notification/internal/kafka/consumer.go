package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type LargeTransaction struct {
	UserID        string    `json:"user_id"`
	TransactionID string    `json:"transaction_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Type          string    `json:"type"`
	Timestamp     time.Time `json:"timestamp"`
}

type Consumer struct {
	reader  *kafka.Reader
	handler func(context.Context, *LargeTransaction) error
}

func NewConsumer(brokers, groupID, topic string, handler func(context.Context, *LargeTransaction) error) (*Consumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokers},
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &Consumer{
		reader:  reader,
		handler: handler,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				continue
			}

			var tx LargeTransaction
			if err := json.Unmarshal(msg.Value, &tx); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			if err := c.handler(ctx, &tx); err != nil {
				log.Printf("Error handling transaction: %v", err)
				continue
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("Error committing message: %v", err)
			}
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
