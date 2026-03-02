package shutdown

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type GracefulShutdown struct {
	signals chan os.Signal
	timeout time.Duration
}

func NewGracefulShutdown(timeout time.Duration) *GracefulShutdown {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	return &GracefulShutdown{
		signals: signals,
		timeout: timeout,
	}
}

func (g *GracefulShutdown) Wait() {
	<-g.signals
	log.Println("Shutdown signal received")
}

func (g *GracefulShutdown) WithTimeout() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	go func() {
		<-g.signals
		log.Println("Shutdown signal received, starting graceful shutdown...")
		cancel()
	}()
	return ctx
}
