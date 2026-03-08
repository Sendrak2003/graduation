package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	_ "gw-currency-wallet/docs"
	"gw-currency-wallet/internal/config"
	"gw-currency-wallet/internal/grpc"
	"gw-currency-wallet/internal/handler"
	"gw-currency-wallet/internal/kafka"
	"gw-currency-wallet/internal/middleware"
	"gw-currency-wallet/internal/repository"
	"gw-currency-wallet/internal/service"
	"gw-currency-wallet/internal/utils/auth"
)

// @title Wallet API
// @version 1.0
// @description API для управления кошельками
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Fatal("db connection failed", zap.Error(err))
	}
	defer db.Close()

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(time.Minute * 10)

	if err := db.Ping(); err != nil {
		logger.Fatal("db ping failed", zap.Error(err))
	}

	logger.Info("Database connected successfully")

	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "kafka:9092"
	}

	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "large-transactions"
	}

	kafkaProducer := kafka.NewProducer(kafkaBrokers, kafkaTopic, logger)
	defer kafkaProducer.Close()

	logger.Info("Kafka producer initialized",
		zap.String("brokers", kafkaBrokers),
		zap.String("topic", kafkaTopic))

	exchangerAddress := os.Getenv("EXCHANGER_GRPC_ADDRESS")
	if exchangerAddress == "" {
		exchangerAddress = "gw-exchanger:50051"
	}

	grpcClient, err := grpc.NewExchangeClient(exchangerAddress, logger)
	if err != nil {
		logger.Warn("failed to create grpc client, exchange functionality will be disabled", zap.Error(err))
		grpcClient = nil
	} else {
		logger.Info("gRPC client initialized", zap.String("address", exchangerAddress))
	}

	authCfg := config.LoadAuth()
	jwtManager := auth.New(
		authCfg.Secret,
		authCfg.AccessTTL,
		authCfg.RefreshTTL,
	)

	repos := repository.NewRepositories(db)

	services := service.NewServices(repos, jwtManager, kafkaProducer, grpcClient, logger)

	authHandler := handler.NewAuthHandler(services.UserService, jwtManager, logger)
	walletHandler := handler.NewWalletHandler(services.WalletService, logger)
	exchangeHandler := handler.NewExchangeHandler(services.ExchangeService, logger)

	if grpcClient != nil {
		defer grpcClient.Close()
	}

	mainHandler := handler.NewHandler(walletHandler, authHandler, exchangeHandler, jwtManager)

	router := gin.Default()

	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.NewPanicRecoveryMiddleware(logger))

	router.GET("/health", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	handler.RegisterRoutes(router, mainHandler)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.Info("Server starting", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server start failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited gracefully")
}
