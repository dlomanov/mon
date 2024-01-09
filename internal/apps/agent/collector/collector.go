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
	sb := strings.Builder{}

	failed := 0
	for _, v := range c.metrics {
		model := apimodels.MapToModel(v)
		data, err := compressJson(model)
		if err != nil {
			failed++
			writeerr(&sb, err.Error())
			continue
		}

		resp, err := c.client.
			R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			SetHeader("Accept-Encoding", "gzip").
			SetBody(data).
			Post("/update/")

		if !resp.IsSuccess() {
			failed++
			writeerr(&sb, "failed request")
			continue
		}

		if err != nil {
			failed++
			writeerr(&sb, err.Error())
		}
	}

	if failed == 0 {
		c.logger.Printf("%d metrics reported\n", len(c.metrics))
		return
	}

	c.logger.Printf("%d metrics reported, %d failed\n%v", len(c.metrics)-failed, failed, sb.String())
}

func compressJson(model apimodels.Metric) ([]byte, error) {
	data, err := json.Marshal(model)
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

func writeerr(sb *strings.Builder, err string) {
	sb.WriteString(" - ")
	sb.WriteString(err)
	sb.WriteString("\n")
}
