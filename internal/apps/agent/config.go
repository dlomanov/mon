package agent

import (
	"github.com/dlomanov/mon/internal/apps/agent/collector"
)

type Config struct {
	CollectorConfig collector.Config
	LogLevel        string
	Addr            string
	GRPCAddr        string
	HashKey         string
	RateLimit       uint64
	PublicKeyPath   string
}
