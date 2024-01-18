package main

import (
	"flag"
	"github.com/caarlos0/env/v10"
	"github.com/dlomanov/mon/internal/apps/server"
	"time"
)

type rawConfig struct {
	Addr            string `env:"ADDRESS"`
	LogLevel        string `env:"LOG_LEVEL"`
	StoreInterval   uint64 `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

const defaultDSN = "host=localhost port=5432 user=postgres password=1 dbname=postgres sslmode=disable"

func getConfig() server.Config {
	raw := rawConfig{}

	flag.StringVar(&raw.Addr, "a", "localhost:8080", "server address")
	flag.StringVar(&raw.LogLevel, "l", "info", "log level")
	flag.Uint64Var(&raw.StoreInterval, "STORE_INTERVAL", 300, "store interval in seconds")
	flag.StringVar(&raw.FileStoragePath, "FILE_STORAGE_PATH", "/tmp/metrics-db.json", "file storage path")
	flag.BoolVar(&raw.Restore, "RESTORE", true, "restore metrics from file at server start")
	flag.StringVar(&raw.DatabaseDSN, "d", defaultDSN, "database DSN")
	flag.Parse()

	err := env.Parse(&raw)
	if err != nil {
		panic(err)
	}

	return server.Config{
		Addr:            raw.Addr,
		LogLevel:        raw.LogLevel,
		StoreInterval:   time.Duration(raw.StoreInterval) * time.Second,
		FileStoragePath: raw.FileStoragePath,
		Restore:         raw.Restore,
		DatabaseDSN:     raw.DatabaseDSN,
	}
}
