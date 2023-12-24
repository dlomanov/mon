package main

import (
	"github.com/dlomanov/mon/internal/apps/server"
)

func main() {
	cfg := getConfig()
	err := server.Run(cfg)
	if err != nil {
		panic(err)
	}
}
