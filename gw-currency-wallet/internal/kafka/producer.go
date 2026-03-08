package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type LargeTransactionMessage struct {
	UserID          string    `json:"user_id"`
	TransactionID   string    `json:"transaction_id"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	FromCurrency    string    `json:"from_currency"`
	ToCurrency      string    `json:"to_currency"`
	ExchangedAmount float64   `json:"exchanged_amount"`
	Rate            float64   `json:"rate"`
	Type            string    `json:"type"`
	Timestamp       time.Time `json:"timestamp"`
}

type Producer struct {
	writer *kafka.Writer
	logger *zap.Logger
}

func NewProducer(brokers, topic string, logger *zap.Logger) *Producer {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           kafka.RequireOne,
		Async:                  false,
		AllowAutoTopicCreation: true,
	}

	return &Producer{
		writer: writer,
		logger: logger,
	}
}

func (p *Producer) SendTransaction(ctx context.Context, msg *LargeTransactionMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		p.logger.Error("failed to marshal transaction",
			zap.String("transaction_id", msg.TransactionID),
			zap.Error(err))
		return err
	}

	kafkaMsg := kafka.Message{
		Key:   []byte(msg.TransactionID),
		Value: data,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, kafkaMsg); err != nil {
		p.logger.Error("failed to send message to kafka",
			zap.String("transaction_id", msg.TransactionID),
			zap.Error(err))
		return err
	}

	p.logger.Info("large transaction sent to kafka",
		zap.String("transaction_id", msg.TransactionID),
		zap.String("user_id", msg.UserID),
		zap.Float64("amount", msg.Amount),
		zap.String("type", msg.Type))

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
