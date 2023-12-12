package collector

import (
	"fmt"
	"github.com/dlomanov/mon/internal/entities/metrics"
	"github.com/dlomanov/mon/internal/entities/metrics/counter"
	"github.com/dlomanov/mon/internal/entities/metrics/gauge"
	"github.com/go-resty/resty/v2"
	"log"
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

	addr := c.addr
	if !strings.HasPrefix(addr, "http") { // ensure protocol schema
		addr = "http://" + addr
	}

	client := resty.New()
	client.SetBaseURL(addr)

	for key, v := range c.metrics {
		mtype, name, value := v.Deconstruct()
		_, err := client.
			R().
			SetPathParam("type", mtype).
			SetPathParam("name", name).
			SetPathParam("value", value).
			Post("/update/{type}/{name}/{value}")

		if err != nil {
			sb.WriteString(fmt.Sprintf(" - %s: failed:\n   %v\n", key, err))
			continue
		}

		sb.WriteString(fmt.Sprintf(" - %s: ok\n", key))
	}
	sb.WriteRune('\n')
	c.logger.Print(sb.String())
}
