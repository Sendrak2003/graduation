package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LargeTransactionDoc struct {
	UserID        string    `bson:"user_id"`
	TransactionID string    `bson:"transaction_id"`
	Amount        float64   `bson:"amount"`
	Currency      string    `bson:"currency"`
	Type          string    `bson:"type"`
	Timestamp     time.Time `bson:"timestamp"`
	ProcessedAt   time.Time `bson:"processed_at"`
}

type MongoRepository struct {
	collection *mongo.Collection
}

func NewMongoRepository(uri, database string) (*MongoRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	collection := client.Database(database).Collection("large_transactions")

	return &MongoRepository{
		collection: collection,
	}, nil
}

func (r *MongoRepository) SaveTransaction(ctx context.Context, doc *LargeTransactionDoc) error {
	doc.ProcessedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	return nil
}
