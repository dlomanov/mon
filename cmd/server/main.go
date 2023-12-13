package main

import (
	"fmt"
	"github.com/dlomanov/mon/internal/apps/server"
)

func main() {
	cfg := getConfig()
	fmt.Printf("servers running on %s\n", cfg.Addr)
	err := server.ListenAndServe(cfg.Addr)
	if err != nil {
		panic(err)
	}
}
