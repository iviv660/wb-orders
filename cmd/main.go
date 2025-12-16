package main

import (
	"app/internal/app"
	"app/internal/closer"
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, err := app.NewApp(ctx)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}

	worker, err := application.DIContainer().Worker(ctx)
	if err != nil {
		log.Fatalf("init kafka worker: %v", err)
	}

	errCh := make(chan error, 2)

	go func() { errCh <- worker.Run(ctx) }()
	go func() { errCh <- application.Run(ctx) }()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("stopped with error: %v", err)
		}
		stop()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := closer.CloseAll(shutdownCtx); err != nil {
		log.Printf("closeAll error: %v", err)
	}
}
