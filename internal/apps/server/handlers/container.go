package handlers

import (
	"context"
	"errors"
	"time"

	"github.com/dlomanov/mon/internal/apps/server/logging"
	"github.com/dlomanov/mon/internal/entities"
	"github.com/dlomanov/mon/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"io"
)

const (
	HeaderContentType = "Content-Type"
)

type Storage interface {
	Set(ctx context.Context, metrics ...entities.Metric) error
	Get(ctx context.Context, key entities.MetricsKey) (metric entities.Metric, ok bool, err error)
	All(ctx context.Context) (result []entities.Metric, err error)
}

type Config struct {
	LogLevel        string
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
	DatabaseDSN     string
}

func NewContainer(
	ctx context.Context,
	cfg Config,
) (*Container, error) {
	logger, err := logging.WithLevel(cfg.LogLevel)
	if err != nil {
		return nil, err
	}

	db, err := createDB(ctx, cfg)
	if err != nil {
		return nil, err
	}

	s, err := createStorage(ctx, logger, db, cfg)
	if err != nil {
		return nil, err
	}

	return &Container{
		Context: ctx,
		Logger:  logger,
		DB:      db,
		Storage: s,
	}, nil
}

type Container struct {
	Context context.Context
	DB      *sqlx.DB
	Logger  *zap.Logger
	Storage Storage
}

func (c *Container) Close() (err error) {
	if c.DB != nil {
		err = c.DB.Close()
	}
	if closer, ok := c.Storage.(io.Closer); ok {
		err = errors.Join(err, closer.Close())
	}

	return err
}

func createDB(ctx context.Context, cfg Config) (*sqlx.DB, error) {
	if cfg.DatabaseDSN == "" {
		return nil, nil
	}
	return sqlx.ConnectContext(ctx, "pgx", cfg.DatabaseDSN)
}

func createStorage(
	ctx context.Context,
	logger *zap.Logger,
	db *sqlx.DB,
	cfg Config,
) (Storage, error) {
	switch {
	case db != nil:
		return storage.NewPGStorage(ctx, logger, db)
	case cfg.FileStoragePath != "":
		return createFileStorage(ctx, logger, cfg)
	default:
		return storage.NewMemStorage(), nil
	}
}

func createFileStorage(
	ctx context.Context,
	logger *zap.Logger,
	cfg Config,
) (*storage.FileStorage, error) {
	fs, err := storage.NewFileStorage(logger, storage.FileStorageConfig{
		StoreInterval:   cfg.StoreInterval,
		FileStoragePath: cfg.FileStoragePath,
		Restore:         cfg.Restore,
	})
	if err != nil {
		return nil, err
	}
	go func() {
		err := fs.DumpLoop(ctx)
		if errors.Is(err, context.Canceled) {
			logger.Debug("dump loop cancelled", zap.Error(err))
			return
		}
		if err != nil {
			logger.Error("failed dump loop", zap.Error(err))
		}
	}()
	return fs, nil
}
