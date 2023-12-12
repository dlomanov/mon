package main

import (
	"flag"
	"time"
)

type options struct {
	addr           string
	pollInterval   time.Duration
	reportInterval time.Duration
}

func parseOptions() options {
	addr := flag.String("a", "localhost:8080", "server address")
	poll := flag.Int64("p", 2, "metrics poll interval in seconds")
	report := flag.Int64("r", 10, "metrics poll interval in seconds")
	flag.Parse()
	return options{
		addr:           *addr,
		pollInterval:   time.Duration(int64(time.Second) * *poll),
		reportInterval: time.Duration(int64(time.Second) * *report),
	}
}
