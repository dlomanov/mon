package collector

import (
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
	c.logger.Printf("%d metrics updated", len(c.metrics))
}

func (c *Collector) ReportMetrics() {
	addr := c.addr
	if !strings.HasPrefix(addr, "http") { // ensure protocol schema
		addr = "http://" + addr
	}

	client := resty.New()
	client.SetBaseURL(addr)

	sb := strings.Builder{}

	failed := 0
	for _, v := range c.metrics {
		mtype, name, value := v.Deconstruct()
		_, err := client.
			R().
			SetPathParam("type", mtype).
			SetPathParam("name", name).
			SetPathParam("value", value).
			Post("/update/{type}/{name}/{value}")

		if err == nil {
			continue
		}

		failed++
		sb.WriteString(" - ")
		sb.WriteString(err.Error())
		sb.WriteString("\n")
	}

	if failed == 0 {
		c.logger.Printf("%d metrics reported\n", len(c.metrics))
		return
	}

	c.logger.Printf("%d metrics reported, %d failed\n%v", len(c.metrics)-failed, failed, sb.String())
}
