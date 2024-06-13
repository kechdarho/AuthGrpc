package app

import (
	"context"

	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/cache"
	"sso/internal/config"
	"sso/internal/pkg/storage/sqlite"
	"sso/internal/services/auth"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(ctx context.Context, log *slog.Logger, cfg *config.Config) *App {
	storage, err := sqlite.New(ctx, cfg.Sqlite.Path)
	if err != nil {
		panic(err)
	}
	cache, err := cache.InitCache(ctx, cfg.Cache.DefaultExpiration, cfg.Cache.CleanupInterval)
	if err != nil {
		panic(err)
	}
	authService := auth.New(log, storage, cache, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL, cfg.JWT.SecretKey)

	grpcApp := grpcapp.New(log, authService, cfg.GRPC.Port)
	return &App{GRPCServer: grpcApp}
}
