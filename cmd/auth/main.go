package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sso/internal/app"
	"sso/internal/config"
	"syscall"
)

const (
	envTest = "test"
	envDev  = "dev"
	envProd = "prod"
)

func main() {
	ctx := context.Background()
	cfg := config.Get()
	log := setupLogger(string(cfg.Environment))
	log.Info("starting application",
		slog.String("env", string(cfg.Environment)),
		slog.Any("cfg", cfg),
	)
	application := app.New(ctx, log, cfg)
	go application.GRPCServer.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	sign := <-stop

	log.Info("stopping application", slog.String("signal", sign.String()))
	application.GRPCServer.Stop()
	log.Info("application stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envTest:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}
