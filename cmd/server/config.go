package main

import (
	"flag"
	"github.com/dlomanov/mon/internal/apps/server"
	"os"
)

func getConfig() server.Config {
	cfg := server.Config{}

	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "server address")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.Parse()

	if addr := os.Getenv("ADDRESS"); addr != "" {
		cfg.Addr = addr
	}

	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		cfg.LogLevel = lvl
	}

	return cfg
}
