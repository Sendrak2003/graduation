package grpc

import (
	"context"
	"fmt"
	"time"

	pb "proto-exchange/exchange"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ExchangeClient struct {
	conn   *grpc.ClientConn
	client pb.ExchangeServiceClient
	logger *zap.Logger
}

func NewExchangeClient(address string, logger *zap.Logger) (*ExchangeClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to exchanger service: %w", err)
	}

	client := pb.NewExchangeServiceClient(conn)

	logger.Info("gRPC client connected", zap.String("address", address))

	return &ExchangeClient{
		conn:   conn,
		client: client,
		logger: logger,
	}, nil
}

func (c *ExchangeClient) GetAllExchangeRates(ctx context.Context) (map[string]float64, error) {
	resp, err := c.client.GetExchangeRates(ctx, &pb.Empty{})
	if err != nil {
		c.logger.Error("failed to get all exchange rates", zap.Error(err))
		return nil, fmt.Errorf("failed to get exchange rates: %w", err)
	}

	c.logger.Debug("received exchange rates from grpc", zap.Int("count", len(resp.Rates)))

	rates := make(map[string]float64)
	for k, v := range resp.Rates {
		rates[k] = float64(v)
	}

	return rates, nil
}

func (c *ExchangeClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
