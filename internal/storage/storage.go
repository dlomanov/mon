package storage

type Storage interface {
	Set(key, value string)
	Get(key string) (value string, ok bool)
	All() map[string]string
}

func NewStorage() Storage {
	return &memStorage{
		storage: make(map[string]string),
	}
}
