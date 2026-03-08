package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "proto-exchange/exchange"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"gw-exchanger/internal/config"
	grpcServer "gw-exchanger/internal/grpc"
	"gw-exchanger/internal/service"
	"gw-exchanger/internal/storages/postgres"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	storage, err := postgres.NewPostgresStorage(cfg.GetDBConnectionString())
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer storage.Close()

	logger.Info("Database connected successfully")

	exchangeService := service.NewExchangeService(storage)
	exchangeServer := grpcServer.NewExchangeServer(exchangeService, logger)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	grpcSrv := grpc.NewServer()
	pb.RegisterExchangeServiceServer(grpcSrv, exchangeServer)

	go func() {
		logger.Info("gRPC server starting", zap.String("port", cfg.GRPCPort))
		if err := grpcSrv.Serve(lis); err != nil {
			logger.Fatal("failed to serve", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down gRPC server...")
	grpcSrv.GracefulStop()
	fmt.Println("Server exited gracefully")
}
