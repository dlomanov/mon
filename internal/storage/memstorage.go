package storage

import "sync"

func init() {
	var _ Storage = (*memStorage)(nil)
}

type memStorage struct {
	mu      sync.Mutex
	storage map[string]string
}

func (mem *memStorage) All() map[string]string {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	result := make(map[string]string, len(mem.storage))
	for k, v := range mem.storage {
		result[k] = v
	}

	return result
}

func (mem *memStorage) Set(key, value string) {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	mem.storage[key] = value
}

func (mem *memStorage) Get(key string) (value string, ok bool) {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	v, ok := mem.storage[key]
	return v, ok
}
