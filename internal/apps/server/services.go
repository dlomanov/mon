package server

import (
	"context"
	"github.com/dlomanov/mon/internal/apps/server/logging"
	"github.com/dlomanov/mon/internal/storage"
	"go.uber.org/zap"
)

func configureServices(
	ctx context.Context,
	cfg Config,
) (*serviceContainer, error) {
	logger, err := logging.WithLevel(cfg.LogLevel)
	if err != nil {
		return nil, err
	}

	stg := configureStorage(logger, cfg)
	return &serviceContainer{
		MemStorage: stg,
		Storage:    stg,
		Logger:     logger,
		Context:    ctx,
	}, nil
}

type serviceContainer struct {
	MemStorage *storage.MemStorage
	Storage    storage.Storage
	Logger     *zap.Logger
	Context    context.Context
}

func configureStorage(logger *zap.Logger, cfg Config) *storage.MemStorage {
	return storage.NewMemStorage(
		logger,
		storage.Config{
			StoreInterval:   cfg.StoreInterval,
			FileStoragePath: cfg.FileStoragePath,
			Restore:         cfg.Restore,
		})
}
