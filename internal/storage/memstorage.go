package storage

func init() {
	var _ Storage = (*memStorage)(nil)
}

type memStorage struct {
	storage map[string]string
}

func (mem *memStorage) Set(key, value string) {
	mem.storage[key] = value
}

func (mem *memStorage) Get(key string) (value string, ok bool) {
	v, ok := mem.storage[key]
	return v, ok
}
