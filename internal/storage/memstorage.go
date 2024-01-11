package storage

import (
	"go.uber.org/zap"
	"sync"
	"time"
)

func init() {
	var _ Storage = (*MemStorage)(nil)
}

func NewMemStorage(logger *zap.Logger, config Config) *MemStorage {
	ms := &MemStorage{
		mu:      sync.Mutex{},
		storage: make(map[string]string),
		logger:  logger,
		config:  config,
	}
	_ = load(ms)
	return ms
}

type MemStorage struct {
	mu       sync.Mutex
	storage  map[string]string
	logger   *zap.Logger
	config   Config
	lastDump time.Time
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
	_ = dump(mem, false)
}

func (mem *MemStorage) Get(key string) (value string, ok bool) {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	v, ok := mem.storage[key]
	return v, ok
}

func (mem *MemStorage) Close() error {
	mem.mu.Lock()
	defer mem.mu.Unlock()
	return dump(mem, true)
}
