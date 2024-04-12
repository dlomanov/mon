package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/dlomanov/mon/internal/apps/server"
	"github.com/dlomanov/mon/internal/apps/shared/logging"
)

// main is the entry point of the server application.
// It performs the following steps:
// 1. Loads the application configuration from environment variables or a configuration file.
// 2. Initializes the logger with the specified log level.
// 3. Starts an HTTP server on a specified port for profiling and debugging purposes.
// 4. Runs the server with the loaded configuration, handling incoming requests.
// 5. If an error occurs during the server startup or while running, it logs the error and terminates the application.
// 6. Gracefully shuts down the server upon receiving an interrupt signal (e.g., SIGINT or SIGTERM).
func main() {
	cfg := getConfig()

	go func() { log.Println(http.ListenAndServe("localhost:6061", nil)) }()

	logger, err := logging.WithLevel(cfg.LogLevel)
	if err != nil {
		log.Fatal(err)
	}

	err = server.Run(context.Background(), cfg, logger)
	if err != nil {
		panic(err)
	}
}
