package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/dlomanov/mon/internal/apps/agent"
)

func main() {
	go func() { log.Println(http.ListenAndServe("localhost:6060", nil)) }()

	cfg := getConfig()
	fmt.Printf("agent running...\n%+v\n\n", cfg)
	err := agent.Run(cfg)
	if err != nil {
		panic(err)
	}
}
