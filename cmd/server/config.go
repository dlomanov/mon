package main

import (
	"flag"
	"github.com/caarlos0/env/v10"
	"github.com/dlomanov/mon/internal/apps/server"
	"github.com/dlomanov/mon/internal/apps/server/container"
	"time"
)

type rawConfig struct {
	Addr            string `env:"ADDRESS"`
	LogLevel        string `env:"LOG_LEVEL"`
	StoreInterval   uint64 `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
}

func getConfig() server.Config {
	raw := rawConfig{}

	flag.StringVar(&raw.Addr, "a", "localhost:8080", "server address")
	flag.StringVar(&raw.LogLevel, "l", "info", "log level")
	flag.Uint64Var(&raw.StoreInterval, "i", 300, "store interval in seconds")
	flag.StringVar(&raw.FileStoragePath, "f", "/tmp/metrics-db.json", "file storage path")
	flag.BoolVar(&raw.Restore, "r", true, "restore metrics from file at server start")
	flag.StringVar(&raw.DatabaseDSN, "d", "", "database DSN")
	flag.StringVar(&raw.Key, "k", "", "hashing key")
	flag.Parse()

	err := env.Parse(&raw)
	if err != nil {
		panic(err)
	}

	return server.Config{
		Addr: raw.Addr,
		Config: container.Config{
			LogLevel:        raw.LogLevel,
			StoreInterval:   time.Duration(raw.StoreInterval) * time.Second,
			FileStoragePath: raw.FileStoragePath,
			Restore:         raw.Restore,
			DatabaseDSN:     raw.DatabaseDSN,
			Key:             raw.Key,
		},
	}
}
