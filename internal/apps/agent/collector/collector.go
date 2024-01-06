package collector

import (
	"github.com/dlomanov/mon/internal/apps/apimodels"
	"github.com/dlomanov/mon/internal/entities"
	"github.com/go-resty/resty/v2"
	"log"
	"strings"
)

type Collector struct {
	metrics map[string]entities.Metric
	logger  *log.Logger
	client  *resty.Client
}

func NewCollector(addr string, logger *log.Logger) Collector {
	return Collector{
		client:  createClient(addr),
		metrics: make(map[string]entities.Metric),
		logger:  logger,
	}
}

func createClient(addr string) *resty.Client {
	if !strings.HasPrefix(addr, "http") { // ensure protocol schema
		addr = "http://" + addr
	}
	client := resty.New()
	client.SetBaseURL(addr)
	return client
}

func (c *Collector) UpdateGauge(name string, value float64) {
	key := entities.MetricsKey{Id: name, Type: entities.MetricGauge}
	v := entities.Metric{MetricsKey: key, Value: &value}
	c.metrics[key.String()] = v
}

func (c *Collector) UpdateCounter(name string, value int64) {
	key := entities.MetricsKey{Id: name, Type: entities.MetricCounter}
	keyString := key.String()
	v := entities.Metric{MetricsKey: key, Delta: &value}

	old, ok := c.metrics[keyString]
	if ok {
		*v.Delta += *old.Delta
	}

	c.metrics[keyString] = v
}

func (c *Collector) LogUpdated() {
	c.logger.Printf("%d metrics updated", len(c.metrics))
}

func (c *Collector) ReportMetrics() {
	sb := strings.Builder{}

	failed := 0
	for _, v := range c.metrics {
		model := apimodels.MapToModel(v)
		_, err := c.client.
			R().
			SetBody(model).
			Post("/update/")

		if err != nil {
			failed++
			sb.WriteString(" - ")
			sb.WriteString(err.Error())
			sb.WriteString("\n")
		}
	}

	if failed == 0 {
		c.logger.Printf("%d metrics reported\n", len(c.metrics))
		return
	}

	c.logger.Printf("%d metrics reported, %d failed\n%v", len(c.metrics)-failed, failed, sb.String())
}
