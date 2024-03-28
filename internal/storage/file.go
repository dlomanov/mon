package storage

import (
	"context"
	"sync"
	"time"

	"github.com/dlomanov/mon/internal/entities"
	"github.com/dlomanov/mon/internal/storage/internal/dumper"
	"github.com/dlomanov/mon/internal/storage/internal/mem"
	"go.uber.org/zap"
)

type FileStorageConfig struct {
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
}

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

type FileStorage struct {
	mu       sync.RWMutex
	internal *mem.Storage
	logger   *zap.Logger
	dumper   *dumper.FileDumper
	config   FileStorageConfig
	syncDump bool
	closed   bool
}

func (fs *FileStorage) Close() error {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.dumper.Dump(*fs.internal)
}

func (fs *FileStorage) Get(
	_ context.Context,
	key entities.MetricsKey,
) (metric entities.Metric, ok bool, err error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	metrics := fs.internal.Get(key)
	if len(metrics) == 0 {
		return metric, false, nil
	}

	return metrics[0], true, nil
}

func (fs *FileStorage) All(_ context.Context) (result []entities.Metric, err error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	return fs.internal.All(), nil
}

func (fs *FileStorage) Set(_ context.Context, metrics ...entities.Metric) error {
	fs.mu.Lock()
	fs.internal.Set(metrics...)
	fs.mu.Unlock()

	if fs.syncDump {
		_ = fs.dump()
	}

	return nil
}

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
