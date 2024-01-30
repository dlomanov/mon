package main

import (
	"flag"
	"github.com/caarlos0/env/v10"
	"github.com/dlomanov/mon/internal/apps/agent"
	"time"
)

type rawConfig struct {
	Addr           string `env:"ADDRESS"`
	PollInterval   uint64 `env:"POLL_INTERVAL"`
	ReportInterval uint64 `env:"REPORT_INTERVAL"`
	Key            string `env:"KEY"`
}

func (r rawConfig) toConfig() agent.Config {
	return agent.Config{
		Addr:           r.Addr,
		PollInterval:   time.Duration(r.PollInterval) * time.Second,
		ReportInterval: time.Duration(r.ReportInterval) * time.Second,
		Key:            r.Key,
	}
}

func getConfig() agent.Config {
	raw := rawConfig{}

	flag.StringVar(&raw.Addr, "a", "localhost:8080", "server address")
	flag.Uint64Var(&raw.PollInterval, "p", 2, "metrics poll interval in seconds")
	flag.Uint64Var(&raw.ReportInterval, "r", 10, "metrics poll interval in seconds")
	flag.StringVar(&raw.Key, "k", "", "hashing key")
	flag.Parse()

	err := env.Parse(&raw)
	if err != nil {
		panic(err)
	}

	return raw.toConfig()
}
