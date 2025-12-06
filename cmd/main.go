package main

import (
	"app/internal/adapter/kafka"
	"app/internal/app"
	"app/internal/config"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.Load()
	log.Printf("Starting application with config: HTTP=%s, Kafka=%v", cfg.HTTP.Addr, cfg.Kafka.Brokers)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application, err := app.NewApp(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	orderService := application.DIContainer().OrderService(ctx)

	kafkaReader := kafka.NewKafkaReader(
		cfg.Kafka.Brokers,
		cfg.Kafka.Topic,
		cfg.Kafka.GroupID,
	)
	defer func() {
		if err := kafkaReader.Close(); err != nil {
			log.Printf("Error closing Kafka reader: %v", err)
		}
	}()

	go func() {
		log.Printf("Starting Kafka consumer (topic=%s, group=%s)", cfg.Kafka.Topic, cfg.Kafka.GroupID)
		if err := kafka.RunConsumer(ctx, kafkaReader, orderService); err != nil {
			log.Printf("Kafka consumer stopped: %v", err)
		}
	}()

	go func() {
		log.Printf("Starting HTTP server on %s", cfg.HTTP.Addr)
		if err := application.Run(ctx); err != nil {
			log.Printf("HTTP server stopped: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	cancel()

	if err := application.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down HTTP server: %v", err)
	}

	log.Println("Application stopped")
}
