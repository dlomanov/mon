package main

import (
	"fmt"
	"github.com/dlomanov/mon/internal/apps/agent"
)

func main() {
	cfg := getConfig()
	fmt.Printf("agent running...\n%+v\n\n", cfg)
	err := agent.Run(cfg)
	if err != nil {
		panic(err)
	}
}
