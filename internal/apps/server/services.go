package server

import (
	"context"
	"database/sql"
	"github.com/dlomanov/mon/internal/apps/server/handlers/interfaces"
	"github.com/dlomanov/mon/internal/apps/server/logging"
	"github.com/dlomanov/mon/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func init() {
	var _ interfaces.Storage = (*storage.Storage)(nil)
}

func configureServices(
	ctx context.Context,
	cfg Config,
) (*serviceContainer, error) {
	logger, err := logging.WithLevel(cfg.LogLevel)
	if err != nil {
		return nil, err
	}

	db, err := createDB(cfg)
	if err != nil {
		return nil, err
	}

	stg := createStorage(logger, db, cfg)

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
	MemStorage *storage.Storage
	Storage    interfaces.Storage
}

func createDB(cfg Config) (*sql.DB, error) {
	if cfg.DatabaseDSN == "" {
		return nil, nil
	}
	return sql.Open("pgx", cfg.DatabaseDSN)
}

func createStorage(
	logger *zap.Logger,
	db *sql.DB,
	cfg Config,
) *storage.Storage {
	return storage.NewStorage(logger, db,
		storage.Config{
			StoreInterval:   cfg.StoreInterval,
			FileStoragePath: cfg.FileStoragePath,
			Restore:         cfg.Restore,
		})
}
