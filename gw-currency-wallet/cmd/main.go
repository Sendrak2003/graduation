package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "gw-currency-wallet/docs"
	"gw-currency-wallet/internal/config"
	"gw-currency-wallet/internal/handler"
	httpHandler "gw-currency-wallet/internal/handler/http"
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
		log.Fatalf("db connection failed: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(time.Minute * 10)

	if err := db.Ping(); err != nil {
		log.Fatalf("db ping failed: %v", err)
	}

	log.Println("Database connected successfully")

	authCfg := config.LoadAuth()
	jwtManager := auth.New(
		authCfg.Secret,
		authCfg.AccessTTL,
		authCfg.RefreshTTL,
	)

	repos := repository.NewRepositories(db)

	services := service.NewServices(repos, jwtManager)

	authHandler := handler.NewAuthHandler(services.UserService, jwtManager)
	walletHandler := httpHandler.NewWalletHandler(services.WalletService)
	mainHandler := httpHandler.NewHandler(walletHandler, authHandler, jwtManager)

	// Настройка роутера
	router := gin.Default()

	router.Use(middleware.NewPanicRecoveryMiddleware(nil))

	router.GET("/health", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	httpHandler.RegisterRoutes(router, mainHandler)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server start failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
