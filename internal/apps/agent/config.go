package agent

import "time"

type Config struct {
	Addr           string
	PollInterval   time.Duration
	ReportInterval time.Duration
	Key            string
}
