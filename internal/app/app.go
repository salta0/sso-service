package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/services/auth"
	"sso/internal/storage/sqlite"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, storagePath string, tokenTTL time.Duration) *App {
	// TODO: init storage
	// TODO: init auth service
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}
	auth := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, auth, grpcPort)

	return &App{grpcApp}
}
