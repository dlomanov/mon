package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/dlomanov/mon/internal/apps/server"
)

func main() {
	cfg := getConfig()

	go func() { log.Println(http.ListenAndServe("localhost:6061", nil)) }()

	err := server.Run(cfg)
	if err != nil {
		panic(err)
	}
}
