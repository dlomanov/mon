package storage

import (
	"context"
	"github.com/dlomanov/mon/internal/apps/server/usecases"
	"github.com/dlomanov/mon/internal/infra/storage/internal/dumper"
	"github.com/dlomanov/mon/internal/infra/storage/internal/mem"
	"sync"
	"time"

	"github.com/dlomanov/mon/internal/entities"
	"go.uber.org/zap"
)

var _ usecases.Storage = (*FileStorage)(nil)

type (
	// FileStorageConfig holds the configuration for the FileStorage.
	// It includes settings for the store interval, file storage path,
	// and whether to restore metrics from storage on startup
	FileStorageConfig struct {
		StoreInterval   time.Duration // StoreInterval defines the interval at which metrics are stored.
		FileStoragePath string        // FileStoragePath is the path to the directory where metrics are stored in file storage.
		Restore         bool          // Restore indicates whether to restore metrics from storage on startup.
	}

	// FileStorage is a storage system that uses files for persistence.
	// It provides methods for storing, retrieving, and managing metrics.
	// FileStorage uses an in-memory storage for quick access and periodically
	// dumps the data to files for persistence.
	FileStorage struct {
		mu       sync.RWMutex
		internal *mem.Storage
		logger   *zap.Logger
		dumper   *dumper.FileDumper
		config   FileStorageConfig
		syncDump bool
		closed   bool
	}
)

// NewFileStorage creates a new FileStorage instance with the given configuration.
// It initializes the storage with data from the file if Restore is true.
// Returns an error if the storage cannot be initialized.
func NewFileStorage(logger *zap.Logger, config FileStorageConfig) (*FileStorage, error) {
	fs := &FileStorage{
		mu:       sync.RWMutex{},
		internal: mem.NewStorage(),
		logger:   logger,
		config:   config,
		syncDump: config.StoreInterval == 0,
		dumper:   dumper.NewFileDumper(logger, config.FileStoragePath),
		closed:   false,
	}

	err := fs.load()
	if err != nil {
		return nil, err
	}

	return fs, nil
}

// Close ensures that any pending operations are completed and resources are released.
// It attempts to dump the current state of the in-memory storage to the file system.
// Returns an error if the dump operation fails.
func (fs *FileStorage) Close() error {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.dumper.Dump(*fs.internal)
}

// Get retrieves a metric by its key from the FileStorage.
// Returns the metric and a boolean indicating if the metric was found, or an error if the operation fails.
func (fs *FileStorage) Get(_ context.Context, key entities.MetricsKey) (metric entities.Metric, ok bool, err error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	metrics := fs.internal.Get(key)
	if len(metrics) == 0 {
		return metric, false, nil
	}

	return metrics[0], true, nil
}

// All retrieves all metrics stored in the FileStorage.
// Returns a slice of metrics or an error if the operation fails.
func (fs *FileStorage) All(_ context.Context) ([]entities.Metric, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	return fs.internal.All(), nil
}

// Set sets one or more metrics in the FileStorage.
// Returns an error if the operation fails.
func (fs *FileStorage) Set(_ context.Context, metrics ...entities.Metric) error {
	fs.mu.Lock()
	fs.internal.Set(metrics...)
	fs.mu.Unlock()

	if fs.syncDump {
		_ = fs.dump()
	}

	return nil
}

// DumpLoop starts a loop that periodically dumps the in-memory storage to the file system.
// The loop runs until the provided context is canceled.
// Returns an error if the dump operation fails or if the context is canceled.
func (fs *FileStorage) DumpLoop(ctx context.Context) error {
	if fs.syncDump {
		return nil
	}

	d := fs.config.StoreInterval
	for {
		select {
		case <-ctx.Done():
			fs.logger.Debug("dump loop cancelled", zap.Error(ctx.Err()))
			return ctx.Err()
		case <-time.After(d):
		}

		if err := fs.dump(); err != nil {
			fs.logger.Error("failed dump loop", zap.Error(err))
			return err
		}
	}
}

func (fs *FileStorage) load() error {
	if !fs.config.Restore {
		fs.logger.Debug("load disabled")
		return nil
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()

	err := fs.dumper.Load(fs.internal)
	if err != nil {
		fs.logger.Error("storage loading failed", zap.Error(err))
		return err
	}

	return nil
}

func (fs *FileStorage) dump() error {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	if err := fs.dumper.Dump(*fs.internal); err != nil {
		fs.logger.Error("dump failed", zap.Error(err))
		return err
	}

	return nil
}
