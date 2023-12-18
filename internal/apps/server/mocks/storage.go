package mocks

import (
	"github.com/dlomanov/mon/internal/storage"
	"sync"
)

func init() {
	var _ storage.Storage = (*mockStorage)(nil)
}

func NewStorage() storage.Storage {
	return &mockStorage{
		mu:      sync.Mutex{},
		storage: make(map[string]string),
	}
}

type mockStorage struct {
	mu      sync.Mutex
	storage map[string]string
}

func (m *mockStorage) All() map[string]string {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make(map[string]string, len(m.storage))
	for k, v := range m.storage {
		result[k] = v
	}

	return result
}

func (m *mockStorage) Set(key, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.storage[key] = value
}

func (m *mockStorage) Get(key string) (value string, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	v, ok := m.storage[key]
	return v, ok
}
