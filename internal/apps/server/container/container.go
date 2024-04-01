package container

import (
	"context"
	"errors"
	"io"

	"github.com/dlomanov/mon/internal/entities"
	"github.com/dlomanov/mon/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type (
	// Container is a struct that holds the application's context (database connection, logger, configuration, etc).
	// It serves as a central place for managing dependencies and
	// configuration across the application.
	Container struct {
		Context context.Context
		DB      *sqlx.DB
		Logger  *zap.Logger
		Storage Storage
		Config  Config
	}

	// Storage is an interface that defines methods for storing and retrieving metrics.
	// Implementations of this interface can use different storage backends, such as
	// in-memory storage, file storage, or a database.
	Storage interface {
		Set(ctx context.Context, metrics ...entities.Metric) error
		Get(ctx context.Context, key entities.MetricsKey) (metric entities.Metric, ok bool, err error)
		All(ctx context.Context) (result []entities.Metric, err error)
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

	return &Container{
		Context: ctx,
		Logger:  logger,
		DB:      db,
		Storage: s,
		Config:  cfg,
	}, nil
}

// Close releases any resources held by the container, such as the database connection.
// It should be called when the application is shutting down to ensure
// that all resources are properly released.
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
