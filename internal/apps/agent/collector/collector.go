package collector

import (
	"fmt"
	"github.com/dlomanov/mon/internal/entities/metrics"
	"github.com/dlomanov/mon/internal/entities/metrics/counter"
	"github.com/dlomanov/mon/internal/entities/metrics/gauge"
	"io"
	"log"
	"net/http"
	"strings"
)

type Collector struct {
	metrics map[string]metrics.Metric
	logger  *log.Logger
	addr    string
}

func NewCollector(addr string, logger *log.Logger) Collector {
	return Collector{
		addr:    addr,
		metrics: make(map[string]metrics.Metric),
		logger:  logger,
	}
}

func (c *Collector) UpdateGauge(name string, value float64) {
	v := gauge.Metric{Name: name, Value: value}
	c.metrics[v.Key()] = v
}

func (c *Collector) UpdateCounter(name string, value int64) {
	v := counter.Metric{Name: name, Value: value}
	key := v.Key()
	old, ok := c.metrics[key]
	if ok {
		v.Value += (old.(counter.Metric)).Value
	}
	c.metrics[key] = v
}

func (c *Collector) Updated() {
	sb := strings.Builder{}
	sb.WriteString("METRICS UPDATED:\n")
	for key := range c.metrics {
		sb.WriteString(fmt.Sprintf("- %s\n", key))
	}
	sb.WriteRune('\n')
	c.logger.Print(sb.String())
}

func (c *Collector) ReportMetrics() {
	sb := strings.Builder{}
	sb.WriteString("REPORTING:\n")
	for key, v := range c.metrics {
		mtype, name, value := v.Deconstruct()

		requestURL := fmt.Sprintf("%s/update/%s/%s/%s", c.addr, mtype, name, value)
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
	c.logger.Print(sb.String())
}
