package main

import (
	"fmt"
	"github.com/dlomanov/mon/internal/apps/server"
)

func main() {
	opt := parseOptions()
	fmt.Printf("servers running on %s\n", opt.addr)
	err := server.ListenAndServe(opt.addr)
	if err != nil {
		panic(err)
	}
}
