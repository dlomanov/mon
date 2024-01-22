package collector

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
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
	key := entities.MetricsKey{Name: name, Type: entities.MetricGauge}
	v := entities.Metric{MetricsKey: key, Value: &value}
	c.metrics[key.String()] = v
}

func (c *Collector) UpdateCounter(name string, value int64) {
	key := entities.MetricsKey{Name: name, Type: entities.MetricCounter}
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
	if len(c.metrics) == 0 {
		return
	}

	data := make([]apimodels.Metric, 0, len(c.metrics))
	for _, v := range c.metrics {
		model := apimodels.MapToModel(v)
		data = append(data, model)
	}

	dataJSON, err := compressJSON(data)
	if err != nil {
		c.logger.Println("failed to marshal and compress metrics")
		return
	}

	_, err = c.client.
		R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(dataJSON).
		Post("/updates/")

	if err != nil {
		c.logger.Println("reporting metrics failed")
		return
	}

	c.logger.Println("metrics reported")
}

func compressJSON(models []apimodels.Metric) ([]byte, error) {
	data, err := json.Marshal(models)
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	cw := gzip.NewWriter(&buf)

	_, err = cw.Write(data)
	if err != nil {
		return nil, err
	}
	err = cw.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
