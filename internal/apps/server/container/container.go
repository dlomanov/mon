package container

import (
	"context"
	"errors"
	"github.com/dlomanov/mon/internal/apps/server/usecases"
	"github.com/dlomanov/mon/internal/infra/services/encrypt"
	storage2 "github.com/dlomanov/mon/internal/infra/storage"
	"io"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type (
	// Container is a struct that holds the application's context (database connection, logger, configuration, etc).
	// It serves as a central place for managing dependencies and
	// configuration across the application.
	Container struct {
		Config        Config
		DB            *sqlx.DB
		Logger        *zap.Logger
		Dec           *encrypt.Decryptor
		MetricUseCase *usecases.MetricUseCase
		storage       usecases.Storage
	}
)

// NewContainer creates a new application container with the provided configuration.
// It initializes the application dependencies based on the configuration.
func NewContainer(
	ctx context.Context,
	logger *zap.Logger,
	cfg Config,
) (*Container, error) {
	db, err := createDB(ctx, cfg)
	if err != nil {
		return nil, err
	}

	s, err := createStorage(ctx, logger, db, cfg)
	if err != nil {
		return nil, err
	}

	dec, err := createDecryptor(cfg.PrivateKeyPath)
	if err != nil {
		return nil, err
	}

	metricUC := usecases.NewMetricUseCase(s)

	return &Container{
		Config:        cfg,
		Logger:        logger,
		DB:            db,
		Dec:           dec,
		MetricUseCase: metricUC,
		storage:       s,
	}, nil
}

// Close releases any resources held by the container, such as the database connection.
// It should be called when the application is shutting down to ensure
// that all resources are properly released.
func (c *Container) Close() {
	if closer, ok := c.storage.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			c.Logger.Error("failed to close storage", zap.Error(err))
		}
	}
	if c.DB != nil {
		if err := c.DB.Close(); err != nil {
			c.Logger.Error("failed to close DB", zap.Error(err))
		}
	}
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
) (usecases.Storage, error) {
	switch {
	case db != nil:
		return storage2.NewPGStorage(ctx, logger, db)
	case cfg.FileStoragePath != "":
		return createFileStorage(ctx, logger, cfg)
	default:
		return storage2.NewMemStorage(), nil
	}
}

func createFileStorage(
	ctx context.Context,
	logger *zap.Logger,
	cfg Config,
) (*storage2.FileStorage, error) {
	fs, err := storage2.NewFileStorage(logger, storage2.FileStorageConfig{
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

func createDecryptor(keyPath string) (*encrypt.Decryptor, error) {
	if keyPath == "" {
		return nil, nil
	}
	privateKey, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	return encrypt.NewDecryptor(privateKey)
}
