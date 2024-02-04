package collector

import "time"

type Config struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
}
