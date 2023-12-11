package main

import (
	"github.com/dlomanov/mon/internal/apps/agent"
	"time"
)

const (
	addr           = "http://localhost:8080"
	pollInterval   = time.Second * 2
	reportInterval = time.Second * 10
)

func main() {
	err := agent.Run(addr, pollInterval, reportInterval)
	if err != nil {
		panic(err)
	}
}
