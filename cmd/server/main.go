package main

import (
	"github.com/dlomanov/mon/internal/handlers"
	"github.com/dlomanov/mon/internal/storage"
	"net/http"
)

const port = "8080"

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	db := storage.NewStorage()

	mux := http.NewServeMux()

	// /update/<type>/<name>/<value>
	mux.HandleFunc("/update/", handlers.UpdateHandler(db))

	return http.ListenAndServe(":"+port, mux)
}
