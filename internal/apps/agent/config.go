package agent

import (
	"github.com/dlomanov/mon/internal/apps/agent/collector"
	"github.com/dlomanov/mon/internal/apps/agent/reporter"
)

type Config struct {
	CollectorConfig collector.Config
	ReporterConfig  reporter.Config
	LogLevel        string
}
