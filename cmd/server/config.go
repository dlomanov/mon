package main

import (
	"flag"
	"github.com/caarlos0/env/v10"
	"github.com/dlomanov/mon/internal/apps/server"
)

func getConfig() server.Config {
	cfg := rawConfig{}

	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}

	if cfg.isEmpty() {
		flag.StringVar(&cfg.Addr, "a", "localhost:8080", "server address")
		flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
		flag.Parse()
	}

	return server.Config{
		Addr:     cfg.Addr,
		LogLevel: cfg.LogLevel,
	}
}

type rawConfig struct {
	Addr     string `env:"ADDRESS"`
	LogLevel string `env:"LOG_LEVEL"`
}

func (cfg rawConfig) isEmpty() bool {
	return cfg.Addr == "" || cfg.LogLevel == ""
}
