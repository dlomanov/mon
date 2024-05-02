package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/dlomanov/mon/internal/apps/agent"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

// main is the entry point of the agent application.
// It performs the following steps:
// 1. Loads the application configuration from environment variables or a configuration file.
// 2. Initializes the logger with the specified log level.
// 3. Initializes the metric collector and reporter based on the configuration.
// 4. Runs the agent with the loaded configuration, collecting and reporting metrics.
// 5. If an error occurs during the agent startup or while running, it logs the error and terminates the application.
// 6. Gracefully shuts down the agent upon receiving an interrupt signal (e.g., SIGINT or SIGTERM).
func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n\n", buildCommit)

	go func() { log.Println(http.ListenAndServe("localhost:6060", nil)) }()

	cfg := getConfig()
	err := agent.Run(cfg)
	if err != nil {
		panic(err)
	}
}
