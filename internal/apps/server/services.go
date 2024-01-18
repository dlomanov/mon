package server

import (
	"context"
	"database/sql"
	"github.com/dlomanov/mon/internal/apps/server/logging"
	"github.com/dlomanov/mon/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
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

	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		return nil, err
	}

	stg := configureStorage(logger, cfg)
	return &serviceContainer{
		Context:    ctx,
		Logger:     logger,
		DB:         db,
		MemStorage: stg,
		Storage:    stg,
	}, nil
}

type serviceContainer struct {
	Context    context.Context
	DB         *sql.DB
	Logger     *zap.Logger
	MemStorage *storage.MemStorage
	Storage    storage.Storage
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
