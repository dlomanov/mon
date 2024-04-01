package storage

import (
	"context"
	"sync"

	"github.com/dlomanov/mon/internal/entities"
	"github.com/dlomanov/mon/internal/storage/internal/mem"
)

// MemStorage is an in-memory storage system for metrics.
// It provides methods for storing, retrieving, and managing metrics.
type MemStorage struct {
	internal *mem.Storage
	mu       sync.RWMutex
}

// NewMemStorage creates a new instance of MemStorage.
// It initializes the storage with an empty state.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		internal: mem.NewStorage(),
		mu:       sync.RWMutex{},
	}
}

// Get retrieves a metric by its key from the MemStorage.
// Returns the metric and a boolean indicating if the metric was found, or an error if the operation fails.

func (m *MemStorage) Get(
	ctx context.Context,
	key entities.MetricsKey,
) (
	metric entities.Metric,
	ok bool,
	err error,
) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := m.internal.Get(key)
	if len(metrics) == 0 {
		return metric, false, nil
	}

	return metrics[0], true, nil
}

// All retrieves all metrics stored in the MemStorage.
// Returns a slice of metrics or an error if the operation fails.
func (m *MemStorage) All(ctx context.Context) ([]entities.Metric, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.internal.All(), nil
}

// Set sets one or more metrics in the MemStorage.
// Returns an error if the operation fails.
func (m *MemStorage) Set(ctx context.Context, metrics ...entities.Metric) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.internal.Set(metrics...)
	return nil
}
