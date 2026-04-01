package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"moneyapp/backend/internal/app"
	"moneyapp/backend/internal/config"
)

func main() {
	// Load .env from multiple candidate paths (project root, backend/, cwd)
	_ = godotenv.Load(".env", "../.env", "../../.env")

	cfg := config.MustLoad()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("bootstrap app: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := application.Run(ctx); err != nil {
		log.Fatalf("run app: %v", err)
	}
}
