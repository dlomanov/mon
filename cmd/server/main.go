package main

import (
	"github.com/dlomanov/mon/internal/apps/server"
)

func main() {
	err := server.ListenAndServe(":8080")
	if err != nil {
		panic(err)
	}
}
