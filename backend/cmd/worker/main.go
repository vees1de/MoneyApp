package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"moneyapp/backend/internal/app"
	"moneyapp/backend/internal/config"
)

func main() {
	cfg := config.MustLoad()

	worker, err := app.NewWorker(cfg)
	if err != nil {
		log.Fatalf("bootstrap worker: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := worker.Run(ctx); err != nil {
		log.Fatalf("run worker: %v", err)
	}
}
