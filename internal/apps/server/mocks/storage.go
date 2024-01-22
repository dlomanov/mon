package mocks

import (
	"github.com/dlomanov/mon/internal/apps/server/handlers"
	"github.com/dlomanov/mon/internal/entities"
	"sync"
)

func init() {
	var _ handlers.Storage = (*MockStorage)(nil)
}

func NewStorage() *MockStorage {
	return &MockStorage{
		internal: make(map[string]entities.Metric),
		mu:       sync.Mutex{},
	}
}

type MockStorage struct {
	internal map[string]entities.Metric
	mu       sync.Mutex
}

func (s *MockStorage) Set(metrics ...entities.Metric) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, v := range metrics {
		s.internal[v.String()] = v
	}
}

func (s *MockStorage) Get(key entities.MetricsKey) (metric entities.Metric, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	metric, ok = s.internal[key.String()]
	return metric, ok
}

func (s *MockStorage) All() []entities.Metric {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]entities.Metric, 0, len(s.internal))
	for _, v := range s.internal {
		result = append(result, v)
	}

	return result
}
