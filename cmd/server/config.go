package main

import (
	"flag"
	"github.com/caarlos0/env/v10"
)

type config struct {
	Addr string `env:"ADDRESS"`
}

func (cfg config) isEmpty() bool {
	return cfg.Addr == ""
}

func getConfig() (cfg config) {
	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}
	if !cfg.isEmpty() {
		return cfg
	}

	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "server address")
	flag.Parse()
	return
}
