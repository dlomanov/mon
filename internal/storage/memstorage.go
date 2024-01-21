package storage

import (
	"context"
	"database/sql"
	"github.com/dlomanov/mon/internal/storage/dumper"
	"go.uber.org/zap"
	"sync"
	"time"
)

func init() {
	var _ Storage = (*MemStorage)(nil)
}

func NewMemStorage(
	logger *zap.Logger,
	db *sql.DB,
	config Config,
) *MemStorage {
	mem := &MemStorage{
		mu:       sync.Mutex{},
		storage:  make(map[string]string),
		logger:   logger,
		config:   config,
		syncDump: config.StoreInterval == 0,
		dumper:   createDumper(logger, db, config),
		closed:   false,
	}
	_ = mem.load()
	return mem
}

func createDumper(
	logger *zap.Logger,
	db *sql.DB,
	config Config,
) dumper.Dumper {
	switch {
	case db != nil:
		logger.Debug("db dumper selected")
		return dumper.NewDBDumper(logger, db)
	case config.FileStoragePath != "":
		logger.Debug("file dumper selected")
		return dumper.NewFileDumper(logger, config.FileStoragePath, config.Restore)
	default:
		logger.Debug("noop dumper selected")
		return dumper.NewNoopDumper()
	}
}

type MemStorage struct {
	mu       sync.Mutex
	storage  map[string]string
	logger   *zap.Logger
	config   Config
	syncDump bool
	dumper   dumper.Dumper
	closed   bool
}

func (mem *MemStorage) All() map[string]string {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	result := make(map[string]string, len(mem.storage))
	for k, v := range mem.storage {
		result[k] = v
	}

	return result
}

func (mem *MemStorage) Set(key, value string) {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	mem.storage[key] = value

	if mem.syncDump {
		_ = mem.dump()
	}
}

func (mem *MemStorage) Get(key string) (value string, ok bool) {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	v, ok := mem.storage[key]
	return v, ok
}

func (mem *MemStorage) Close() error {
	if mem.closed {
		return nil
	}

	mem.mu.Lock()
	mem.closed = true
	mem.mu.Unlock()

	return mem.dump()
}

func (mem *MemStorage) DumpLoop(ctx context.Context) error {
	if mem.syncDump {
		return nil
	}

	if _, ok := mem.dumper.(dumper.NoopDumper); ok {
		mem.logger.Debug("dump is disabled")
		return nil
	}

	d := mem.config.StoreInterval
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(d):
		}

		if err := mem.dump(); err != nil {
			return err
		}
	}
}

func (mem *MemStorage) load() error {
	return mem.dumper.Load(&mem.storage)
}

func (mem *MemStorage) dump() error {
	return mem.dumper.Dump(mem.storage)
}
