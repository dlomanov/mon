package mocks

import (
	"context"
	"sync"

	"github.com/dlomanov/mon/internal/entities"
)

func NewStorage() *MockStorage {
	return &MockStorage{
		internal: make(map[string]entities.Metric),
		mu:       sync.RWMutex{},
	}
}

type MockStorage struct {
	internal map[string]entities.Metric
	mu       sync.RWMutex
}

func (s *MockStorage) Set(_ context.Context, metrics ...entities.Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, v := range metrics {
		s.internal[v.String()] = v
	}
	return nil
}

func (s *MockStorage) Get(
	_ context.Context,
	key entities.MetricsKey,
) (metric entities.Metric, ok bool, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metric, ok = s.internal[key.String()]
	return metric, ok, nil
}

func (s *MockStorage) All(_ context.Context) ([]entities.Metric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]entities.Metric, 0, len(s.internal))
	for _, v := range s.internal {
		result = append(result, v)
	}

	return result, nil
}
