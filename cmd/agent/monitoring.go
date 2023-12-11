package main

import (
	"fmt"
	"github.com/dlomanov/mon/internal/handlers/metrics"
	"github.com/dlomanov/mon/internal/handlers/metrics/counter"
	"github.com/dlomanov/mon/internal/handlers/metrics/gauge"
	"io"
	"log"
	"net/http"
	"strings"
)

const (
	baseURL = "http://localhost:8080"
)

type Mon struct {
	metrics map[string]metrics.Metric
	logger  *log.Logger
}

func NewMon(logger *log.Logger) Mon {
	return Mon{
		metrics: make(map[string]metrics.Metric),
		logger:  logger,
	}
}

func (m *Mon) UpdateGauge(name string, value float64) {
	v := gauge.Metric{Name: name, Value: value}
	m.metrics[v.Key()] = v
}

func (m *Mon) UpdateCounter(name string, value int64) {
	v := counter.Metric{Name: name, Value: value}
	key := v.Key()
	old, ok := m.metrics[key]
	if ok {
		v.Value += (old.(counter.Metric)).Value
	}
	m.metrics[key] = v
}

func (m *Mon) Updated() {
	sb := strings.Builder{}
	sb.WriteString("METRICS UPDATED:\n")
	for key := range m.metrics {
		sb.WriteString(fmt.Sprintf("- %s\n", key))
	}
	sb.WriteRune('\n')
	m.logger.Print(sb.String())
}

func (m *Mon) ReportMetrics() {
	sb := strings.Builder{}
	sb.WriteString("REPORTING:\n")
	for key, v := range m.metrics {
		mtype, name, value := v.Deconstruct()

		requestURL := fmt.Sprintf("%s/update/%s/%s/%s", baseURL, mtype, name, value)
		res, err := http.Post(requestURL, "text/plain", nil)
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()

		if err != nil {
			sb.WriteString(fmt.Sprintf(" - %s: failed:\n   %v\n", key, err))
		} else {
			sb.WriteString(fmt.Sprintf(" - %s: ok\n", key))
		}
	}
	sb.WriteRune('\n')
	m.logger.Print(sb.String())
}
