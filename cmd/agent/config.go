package main

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v10"
	"time"
)

func init() {
	var _ fmt.Stringer = (*config)(nil)
}

type config struct {
	Addr           string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

func (c config) String() string {
	return fmt.Sprintf(`config:
- ADDRESS:         %s
- POLL_INTERVAL:   %v
- REPORT_INTERVAL: %v`, c.Addr, c.PollInterval, c.ReportInterval)
}

type rawConfig struct {
	Addr           string `env:"ADDRESS"`
	PollInterval   uint64 `env:"POLL_INTERVAL"`
	ReportInterval uint64 `env:"REPORT_INTERVAL"`
}

func (r rawConfig) isEmpty() bool {
	return r.Addr == "" || r.PollInterval == 0 || r.ReportInterval == 0
}

func (r rawConfig) toConfig() config {
	if r.isEmpty() {
		panic("invalid configuration")
	}
	return config{
		Addr:           r.Addr,
		PollInterval:   time.Duration(int64(time.Second) * int64(r.PollInterval)),
		ReportInterval: time.Duration(int64(time.Second) * int64(r.ReportInterval)),
	}
}

func getConfig() config {
	raw := rawConfig{}

	err := env.Parse(&raw)
	if err != nil {
		panic(err)
	}
	if !raw.isEmpty() {
		return raw.toConfig()
	}

	flag.StringVar(&raw.Addr, "a", "localhost:8080", "server address")
	flag.Uint64Var(&raw.PollInterval, "p", 2, "metrics poll interval in seconds")
	flag.Uint64Var(&raw.ReportInterval, "r", 10, "metrics poll interval in seconds")
	flag.Parse()

	return raw.toConfig()
}
