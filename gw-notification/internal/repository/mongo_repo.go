package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

type LargeTransactionDoc struct {
	UserID          string    `bson:"user_id"`
	TransactionID   string    `bson:"transaction_id"`
	Amount          float64   `bson:"amount"`
	Currency        string    `bson:"currency"`
	FromCurrency    string    `bson:"from_currency"`
	ToCurrency      string    `bson:"to_currency"`
	ExchangedAmount float64   `bson:"exchanged_amount"`
	Rate            float64   `bson:"rate"`
	Type            string    `bson:"type"`
	Timestamp       time.Time `bson:"timestamp"`
	CreatedAt       time.Time `bson:"created_at"`
}

type MongoRepository struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

func NewMongoRepository(uri, database string, logger *zap.Logger) (*MongoRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	logger.Info("MongoDB connected successfully")

	collection := client.Database(database).Collection("large_transactions")

	repo := &MongoRepository{
		collection: collection,
		logger:     logger,
	}

	if err := repo.createIndexes(ctx); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return repo, nil
}

func (r *MongoRepository) createIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "transaction_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "timestamp", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}

	r.logger.Info("MongoDB indexes created successfully")
	return nil
}

func (r *MongoRepository) SaveTransaction(ctx context.Context, doc *LargeTransactionDoc) error {
	doc.CreatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			r.logger.Warn("duplicate transaction",
				zap.String("transaction_id", doc.TransactionID))
			return fmt.Errorf("transaction already exists: %w", err)
		}
		r.logger.Error("failed to insert transaction",
			zap.String("transaction_id", doc.TransactionID),
			zap.Error(err))
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	r.logger.Debug("transaction saved to MongoDB",
		zap.String("transaction_id", doc.TransactionID),
		zap.String("user_id", doc.UserID))

	return nil
}
