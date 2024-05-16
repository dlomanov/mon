package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/dlomanov/mon/internal/apps/agent"
	"github.com/dlomanov/mon/internal/apps/agent/collector"
	"gopkg.in/yaml.v2"
)

type rawConfig struct {
	Addr           string `json:"address" env:"ADDRESS"`
	GRPCAddr       string `json:"grpc_address" env:"GRPC_ADDRESS"`
	PollInterval   uint64 `json:"poll_interval" env:"POLL_INTERVAL"`
	ReportInterval uint64 `json:"report_interval" env:"REPORT_INTERVAL"`
	Key            string `json:"key" env:"KEY"`
	RateLimit      uint64 `json:"rate_limit" env:"RATE_LIMIT"`
	LogLevel       string `json:"log_level" env:"LOG_LEVEL"`
	PublicKeyPath  string `json:"crypto_key" env:"CRYPTO_KEY"`
	ConfigPath     string `json:"config" env:"CONFIG"`
}

//go:embed config.json
var configFS embed.FS

func getConfig() agent.Config {
	raw := rawConfig{}
	raw.readDefault()
	raw.readConfig()
	raw.readFlags()
	raw.readEnv()
	raw.print()
	return raw.toConfig()
}

func (r *rawConfig) readDefault() {
	content, err := configFS.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(content, &r)
	if err != nil {
		panic(err)
	}
}

func (r *rawConfig) readConfig() {
	path := new(string)
	*path = ""

	v := flag.Lookup("c")
	if v == nil {
		v = flag.Lookup("config")
	}
	if v != nil {
		*path = v.Value.String()
	}
	cp, ok := os.LookupEnv("CONFIG")
	if ok {
		*path = cp
	}
	if *path == "" {
		return
	}

	content, err := os.ReadFile(*path)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(content, &r)
	if err != nil {
		panic(err)
	}
}

func (r *rawConfig) readFlags() {
	flag.StringVar(&r.Addr, "a", r.Addr, "server address")
	flag.StringVar(&r.GRPCAddr, "grpc_address", r.GRPCAddr, "gRPC-server address")
	flag.Uint64Var(&r.PollInterval, "p", r.PollInterval, "metrics poll interval in seconds")
	flag.Uint64Var(&r.ReportInterval, "r", r.ReportInterval, "metrics report interval in seconds")
	flag.StringVar(&r.Key, "k", r.Key, "hashing key")
	flag.Uint64Var(&r.RateLimit, "l", r.RateLimit, "report rate limit")
	flag.StringVar(&r.LogLevel, "log_level", r.LogLevel, "log level")
	flag.StringVar(&r.PublicKeyPath, "crypto-key", r.PublicKeyPath, "public key PEM path")
	flag.StringVar(&r.ConfigPath, "config", r.ConfigPath, "config path")
	flag.StringVar(&r.ConfigPath, "c", r.ConfigPath, "config path (shorthand)")
	flag.Parse()
}

func (r *rawConfig) readEnv() {
	err := env.Parse(r)
	if err != nil {
		panic(err)
	}
}

func (r *rawConfig) print() error {
	cyaml, err := yaml.Marshal(r)
	if err != nil {
		return err
	}
	fmt.Println(string(cyaml))
	return nil
}

func (r *rawConfig) toConfig() agent.Config {
	return agent.Config{
		CollectorConfig: collector.Config{
			PollInterval:   time.Duration(r.PollInterval) * time.Second,
			ReportInterval: time.Duration(r.ReportInterval) * time.Second,
		},
		Addr:          r.Addr,
		GRPCAddr:      r.GRPCAddr,
		HashKey:       r.Key,
		RateLimit:     r.RateLimit,
		PublicKeyPath: r.PublicKeyPath,
		LogLevel:      r.LogLevel,
	}
}
