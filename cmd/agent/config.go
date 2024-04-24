package main

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/dlomanov/mon/internal/apps/agent"
	"github.com/dlomanov/mon/internal/apps/agent/collector"
	"github.com/dlomanov/mon/internal/apps/agent/reporter"
)

type rawConfig struct {
	Addr           string `env:"ADDRESS"`
	PollInterval   uint64 `env:"POLL_INTERVAL"`
	ReportInterval uint64 `env:"REPORT_INTERVAL"`
	Key            string `env:"KEY"`
	RateLimit      uint64 `env:"RATE_LIMIT"`
	LogLevel       string `env:"LOG_LEVEL"`
	PublicKeyPath  string `env:"CRYPTO_KEY"`
}

func (r rawConfig) toConfig() agent.Config {
	return agent.Config{
		CollectorConfig: collector.Config{
			PollInterval:   time.Duration(r.PollInterval) * time.Second,
			ReportInterval: time.Duration(r.ReportInterval) * time.Second,
		},
		ReporterConfig: reporter.Config{
			Addr:          r.Addr,
			Key:           r.Key,
			RateLimit:     r.RateLimit,
			PublicKeyPath: r.PublicKeyPath,
		},
		LogLevel: r.LogLevel,
	}
}

func getConfig() agent.Config {
	raw := rawConfig{}

	flag.StringVar(&raw.Addr, "a", "localhost:8080", "server address")
	flag.Uint64Var(&raw.PollInterval, "p", 2, "metrics poll interval in seconds")
	flag.Uint64Var(&raw.ReportInterval, "r", 10, "metrics report interval in seconds")
	flag.StringVar(&raw.Key, "k", "", "hashing key")
	flag.Uint64Var(&raw.RateLimit, "l", 2, "report rate limit")
	flag.StringVar(&raw.LogLevel, "log_level", "info", "log level")
	flag.StringVar(&raw.PublicKeyPath, "crypto-key", "", "public key PEM path")
	flag.Parse()

	err := env.Parse(&raw)
	if err != nil {
		panic(err)
	}

	return raw.toConfig()
}
