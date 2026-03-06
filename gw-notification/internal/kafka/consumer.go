package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type LargeTransaction struct {
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

type Consumer struct {
	reader       *kafka.Reader
	handler      func(context.Context, *LargeTransaction) error
	maxRetries   int
	retryDelay   time.Duration
	logger       *zap.Logger
	batchSize    int
	batchTimeout time.Duration
}

func NewConsumer(brokers, groupID, topic string, handler func(context.Context, *LargeTransaction) error, logger *zap.Logger) (*Consumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokers},
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	return &Consumer{
		reader:       reader,
		handler:      handler,
		maxRetries:   3,
		retryDelay:   time.Second,
		logger:       logger,
		batchSize:    50,
		batchTimeout: 3 * time.Second,
	}, nil
}

func (c *Consumer) processWithRetry(ctx context.Context, tx *LargeTransaction) error {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			delay := c.retryDelay * time.Duration(attempt)
			c.logger.Info("retrying message processing",
				zap.String("transaction_id", tx.TransactionID),
				zap.Int("attempt", attempt),
				zap.Duration("delay", delay))

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := c.handler(ctx, tx)
		if err == nil {
			if attempt > 0 {
				c.logger.Info("message processed successfully after retry",
					zap.String("transaction_id", tx.TransactionID),
					zap.Int("attempt", attempt))
			}
			return nil
		}

		lastErr = err
		c.logger.Warn("message processing failed",
			zap.String("transaction_id", tx.TransactionID),
			zap.Int("attempt", attempt),
			zap.Error(err))
	}

	c.logger.Error("message processing failed after all retries",
		zap.String("transaction_id", tx.TransactionID),
		zap.Int("max_retries", c.maxRetries),
		zap.Error(lastErr))

	return lastErr
}

func (c *Consumer) processBatch(ctx context.Context, batch []*LargeTransaction) error {
	if len(batch) == 0 {
		return nil
	}

	c.logger.Info("processing batch",
		zap.Int("batch_size", len(batch)))

	successCount := 0
	failCount := 0

	for _, tx := range batch {
		if err := c.processWithRetry(ctx, tx); err != nil {
			failCount++
			c.logger.Error("batch item processing failed",
				zap.String("transaction_id", tx.TransactionID),
				zap.Error(err))
		} else {
			successCount++
		}
	}

	c.logger.Info("batch processing completed",
		zap.Int("total", len(batch)),
		zap.Int("success", successCount),
		zap.Int("failed", failCount))

	return nil
}

func (c *Consumer) Start(ctx context.Context) error {
	c.logger.Info("kafka consumer started",
		zap.Int("batch_size", c.batchSize),
		zap.Duration("batch_timeout", c.batchTimeout))

	batch := make([]*LargeTransaction, 0, c.batchSize)
	messages := make([]kafka.Message, 0, c.batchSize)
	ticker := time.NewTicker(c.batchTimeout)
	defer ticker.Stop()

	processBatchAndCommit := func() {
		if len(batch) == 0 {
			return
		}

		if err := c.processBatch(ctx, batch); err != nil {
			c.logger.Error("batch processing error", zap.Error(err))
		}

		if len(messages) > 0 {
			if err := c.reader.CommitMessages(ctx, messages...); err != nil {
				c.logger.Error("failed to commit batch", zap.Error(err))
			} else {
				c.logger.Debug("batch committed", zap.Int("count", len(messages)))
			}
		}

		batch = batch[:0]
		messages = messages[:0]
	}

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("processing remaining batch before shutdown")
			processBatchAndCommit()
			c.logger.Info("kafka consumer stopped")
			return ctx.Err()

		case <-ticker.C:
			c.logger.Debug("batch timeout reached", zap.Int("current_size", len(batch)))
			processBatchAndCommit()
			ticker.Reset(c.batchTimeout)

		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				c.logger.Error("failed to fetch message", zap.Error(err))
				continue
			}

			c.logger.Debug("message received",
				zap.String("topic", msg.Topic),
				zap.Int("partition", msg.Partition),
				zap.Int64("offset", msg.Offset))

			var tx LargeTransaction
			if err := json.Unmarshal(msg.Value, &tx); err != nil {
				c.logger.Error("failed to unmarshal message",
					zap.Error(err),
					zap.ByteString("message", msg.Value))
				continue
			}

			batch = append(batch, &tx)
			messages = append(messages, msg)

			if len(batch) >= c.batchSize {
				c.logger.Debug("batch size reached", zap.Int("size", len(batch)))
				processBatchAndCommit()
				ticker.Reset(c.batchTimeout)
			}
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
