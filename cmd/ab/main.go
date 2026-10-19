package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"AB_system/internal/app"
	"AB_system/internal/configs"
	"AB_system/logger"

	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	_ = godotenv.Load()

	cfg := configs.Load()
	logger.Init(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	a, err := app.New(ctx, cfg)
	if err != nil {
		return err
	}
	return a.Run(ctx)
}
