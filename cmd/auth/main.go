package main

import (
	"AuthGrpc/internal/app"
	"AuthGrpc/internal/app/pprof"
	"AuthGrpc/internal/config"
	"AuthGrpc/internal/lib/profiler"
	"context"
	"log/slog"
	"os"
	"os/signal"
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

	profApplication := pprofapp.New(cfg.ProfileServer.Host, cfg.ProfileServer.Port)
	go profApplication.Start()

	profilerConfig := profiler.ProfilerConfig{
		CPUProfilePath: cfg.Profile.CPUProfilePath,
		MemProfilePath: cfg.Profile.MemProfilePath,
	}
	appProfiler := profiler.NewProfiler(profilerConfig)

	if err := appProfiler.StartCPUProfile(); err != nil {
		panic(err)
	}
	defer appProfiler.StopCPUProfile()

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
