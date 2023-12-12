package main

import (
	"flag"
)

type options struct {
	addr string
}

func parseOptions() (opt options) {
	flag.StringVar(&opt.addr, "a", "localhost:8080", "server address")
	flag.Parse()
	return
}
