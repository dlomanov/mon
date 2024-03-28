package storage

import (
	"context"
	"sync"

	"github.com/dlomanov/mon/internal/entities"
	"github.com/dlomanov/mon/internal/storage/internal/mem"
)

func NewMemStorage() *MemStorage {
	return &MemStorage{
		internal: mem.NewStorage(),
		mu:       sync.RWMutex{},
	}
}

type MemStorage struct {
	internal *mem.Storage
	mu       sync.RWMutex
}

func (m *MemStorage) Get(
	_ context.Context,
	key entities.MetricsKey,
) (metric entities.Metric, ok bool, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := m.internal.Get(key)
	if len(metrics) == 0 {
		return metric, false, nil
	}

	return metrics[0], true, nil
}

func (m *MemStorage) All(_ context.Context) (result []entities.Metric, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.internal.All(), nil
}

func (m *MemStorage) Set(_ context.Context, metrics ...entities.Metric) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.internal.Set(metrics...)
	return nil
}
