package storage

import "time"

type Config struct {
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
}
