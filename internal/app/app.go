package app

import (
	"AuthGrpc/internal/cache/local"
	"context"

	grpcapp "AuthGrpc/internal/app/grpc"
	"AuthGrpc/internal/config"
	"AuthGrpc/internal/pkg/storage/sqlite"
	"AuthGrpc/internal/services/auth"
	"log/slog"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(ctx context.Context, log *slog.Logger, cfg *config.Config) *App {
	storage, err := sqlite.New(ctx, cfg.Sqlite.Path)
	if err != nil {
		panic(err)
	}
	cache, err := local.InitCache(ctx, cfg.Cache.DefaultExpiration, cfg.Cache.CleanupInterval)
	if err != nil {
		panic(err)
	}
	authService := auth.New(log, storage, cache, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL, cfg.JWT.SecretKey)

	grpcApp := grpcapp.New(log, authService, cfg.GRPC.Port)
	return &App{GRPCServer: grpcApp}
}
