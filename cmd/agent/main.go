package main

import (
	"fmt"
	"github.com/dlomanov/mon/internal/apps/agent"
)

func main() {
	opt := parseOptions()
	fmt.Printf("agent running on %s\n", opt.addr)
	err := agent.Run(opt.addr, opt.pollInterval, opt.reportInterval)
	if err != nil {
		panic(err)
	}
}
