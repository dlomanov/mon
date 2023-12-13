package main

import (
	"fmt"
	"github.com/dlomanov/mon/internal/apps/agent"
)

func main() {
	cfg := getConfig()
	fmt.Printf("agent running...\n%s\n\n", cfg)
	err := agent.Run(cfg.Addr, cfg.PollInterval, cfg.ReportInterval)
	if err != nil {
		panic(err)
	}
}
