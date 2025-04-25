package app

import (
	grpcapp "github.com/mjolner36/trade_item_service/internal/app/grpc"
	"github.com/mjolner36/trade_item_service/internal/services/auth"
	"github.com/mjolner36/trade_item_service/internal/storage"
	"log/slog"
	"time"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	storagePort string,
	tokenTTL time.Duration,
) *App {
	//TODO: инициализация storage
	storage, err := storage.New(log, storagePort)
	authService := auth.New(log, storage, storage, storage, tokenTTL) //очень странно выглядит
	//
	grpcApp := grpcapp.New(log, authService, grpcPort)
	return &App{
		GRPCServer: grpcApp,
	}
}
