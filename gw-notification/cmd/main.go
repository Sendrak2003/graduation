package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Println("Notification service starting...")

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:admin@localhost:27017"
	}

	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}

	log.Printf("MongoDB URI: %s", mongoURI)
	log.Printf("Kafka Brokers: %s", kafkaBrokers)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			log.Println("Notification service is running...")
			time.Sleep(30 * time.Second)
		}
	}()

	<-quit
	fmt.Println("Shutting down notification service...")
	fmt.Println("Service exited gracefully")
}
