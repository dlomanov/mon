package storage

import (
	"context"
	"database/sql"
	"github.com/dlomanov/mon/internal/entities"
	"github.com/dlomanov/mon/internal/storage/internal/dumper"
	"github.com/dlomanov/mon/internal/storage/internal/mem"
	"go.uber.org/zap"
	"sync"
	"time"
)

func NewStorage(
	logger *zap.Logger,
	db *sql.DB,
	config Config,
) *Storage {
	s := &Storage{
		mu:       sync.Mutex{},
		internal: mem.NewStorage(),
		logger:   logger,
		config:   config,
		syncDump: config.StoreInterval == 0,
		dumper:   createDumper(logger, db, config),
		closed:   false,
	}
	_ = s.load()
	return s
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

type Storage struct {
	mu       sync.Mutex
	internal *mem.Storage
	logger   *zap.Logger
	config   Config
	syncDump bool
	dumper   dumper.Dumper
	closed   bool
}

func (s *Storage) All() []entities.Metric {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.internal.All()
}

func (s *Storage) Set(metrics ...entities.Metric) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.internal.Set(metrics...)

	if s.syncDump {
		_ = s.dump()
	}
}

func (s *Storage) Get(key entities.MetricsKey) (metric entities.Metric, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	metrics := s.internal.Get(key)
	if len(metrics) == 0 {
		return metric, false
	}

	return metrics[0], true
}

func (s *Storage) Close() error {
	if s.closed {
		return nil
	}

	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()

	return s.dump()
}

func (s *Storage) DumpLoop(ctx context.Context) error {
	if s.syncDump {
		return nil
	}

	if _, ok := s.dumper.(dumper.NoopDumper); ok {
		s.logger.Debug("dump is disabled")
		return nil
	}

	d := s.config.StoreInterval
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(d):
		}

		if err := s.dumpLock(); err != nil {
			return err
		}
	}
}

func (s *Storage) load() error {
	return s.dumper.Load(s.internal)
}

func (s *Storage) dump() error {
	return s.dumper.Dump(*s.internal)
}

func (s *Storage) dumpLock() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.dumper.Dump(*s.internal)
}
