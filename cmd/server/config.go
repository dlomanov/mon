package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/dlomanov/mon/internal/apps/server"
	"github.com/dlomanov/mon/internal/apps/server/container"
	"gopkg.in/yaml.v2"
)

type rawConfig struct {
	Addr            string `json:"address" env:"ADDRESS"`
	GRPCAddr        string `json:"grpc_address" env:"GRPC_ADDRESS"`
	LogLevel        string `json:"log_level" env:"LOG_LEVEL"`
	StoreInterval   uint64 `json:"store_interval" env:"STORE_INTERVAL"`
	FileStoragePath string `json:"file_storage_path" env:"FILE_STORAGE_PATH"`
	Restore         bool   `json:"restore" env:"RESTORE"`
	DatabaseDSN     string `json:"database_dsn" env:"DATABASE_DSN"`
	Key             string `json:"key" env:"KEY"`
	PrivateKeyPath  string `json:"crypto_key" env:"CRYPTO_KEY"`
	ConfigPath      string `json:"config" env:"CONFIG"`
	TrustedSubnet   string `json:"trusted_subnet" env:"TRUSTED_SUBNET"`
}

//go:embed config.json
var configFS embed.FS

func getConfig() server.Config {
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
	flag.StringVar(&r.LogLevel, "l", r.LogLevel, "log level")
	flag.Uint64Var(&r.StoreInterval, "i", r.StoreInterval, "store interval in seconds")
	flag.StringVar(&r.FileStoragePath, "f", r.FileStoragePath, "file storage path")
	flag.BoolVar(&r.Restore, "r", r.Restore, "restore metrics from file at server start")
	flag.StringVar(&r.DatabaseDSN, "d", r.DatabaseDSN, "database DSN")
	flag.StringVar(&r.Key, "k", r.Key, "hashing key")
	flag.StringVar(&r.PrivateKeyPath, "crypto-key", r.PrivateKeyPath, "private key PEM path")
	flag.StringVar(&r.ConfigPath, "config", r.ConfigPath, "config path")
	flag.StringVar(&r.ConfigPath, "c", r.ConfigPath, "config path (shorthand)")
	flag.StringVar(&r.TrustedSubnet, "t", r.TrustedSubnet, "trusted subtnet (CIDR)")
	flag.Parse()
}

func (r *rawConfig) readEnv() {
	err := env.Parse(r)
	if err != nil {
		panic(err)
	}
}

func (r *rawConfig) print() {
	cyaml, err := yaml.Marshal(r)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(cyaml))
}

func (r *rawConfig) toConfig() server.Config {
	cfg := container.Config{
		LogLevel:        r.LogLevel,
		StoreInterval:   time.Duration(r.StoreInterval) * time.Second,
		FileStoragePath: r.FileStoragePath,
		Restore:         r.Restore,
		DatabaseDSN:     r.DatabaseDSN,
		Key:             r.Key,
		Addr:            r.Addr,
		GRPCAddr:        r.GRPCAddr,
		PrivateKeyPath:  r.PrivateKeyPath,
	}

	if r.TrustedSubnet != "" {
		_, subnet, err := net.ParseCIDR(r.TrustedSubnet)
		if err != nil {
			panic(err)
		}
		cfg.TrustedSubnet = subnet
	}

	return server.Config{Config: cfg}
}
