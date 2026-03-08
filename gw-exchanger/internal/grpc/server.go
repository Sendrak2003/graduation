package grpc

import (
	"context"
	"gw-exchanger/internal/service"
	pb "proto-exchange/exchange"

	"go.uber.org/zap"
)

type ExchangeServer struct {
	pb.UnimplementedExchangeServiceServer
	service service.ExchangeServiceInterface
	logger  *zap.Logger
}

func NewExchangeServer(service service.ExchangeServiceInterface, logger *zap.Logger) *ExchangeServer {
	return &ExchangeServer{
		service: service,
		logger:  logger,
	}
}

func (s *ExchangeServer) GetExchangeRateForCurrency(ctx context.Context, req *pb.CurrencyRequest) (*pb.ExchangeRateResponse, error) {
	s.logger.Info("GetExchangeRateForCurrency called",
		zap.String("from", req.FromCurrency),
		zap.String("to", req.ToCurrency))

	rate, err := s.service.GetExchangeRate(ctx, req.FromCurrency, req.ToCurrency)
	if err != nil {
		s.logger.Error("failed to get exchange rate",
			zap.String("from", req.FromCurrency),
			zap.String("to", req.ToCurrency),
			zap.Error(err))
		return nil, err
	}

	return &pb.ExchangeRateResponse{
		FromCurrency: req.FromCurrency,
		ToCurrency:   req.ToCurrency,
		Rate:         float32(rate),
	}, nil
}

func (s *ExchangeServer) GetExchangeRates(ctx context.Context, req *pb.Empty) (*pb.ExchangeRatesResponse, error) {
	s.logger.Info("GetExchangeRates called")

	rates, err := s.service.GetAllExchangeRates(ctx)
	if err != nil {
		s.logger.Error("failed to get all exchange rates", zap.Error(err))
		return nil, err
	}

	ratesFloat32 := make(map[string]float32)
	for k, v := range rates {
		ratesFloat32[k] = float32(v)
	}

	return &pb.ExchangeRatesResponse{
		Rates: ratesFloat32,
	}, nil
}
