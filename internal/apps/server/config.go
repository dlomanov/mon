package server

import "time"

type Config struct {
	Addr            string
	LogLevel        string
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
	DatabaseDSN     string
}
