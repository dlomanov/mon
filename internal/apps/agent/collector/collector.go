package collector

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/dlomanov/mon/internal/apps/shared/apimodels"
	"github.com/dlomanov/mon/internal/apps/shared/hashing"
	"github.com/dlomanov/mon/internal/entities"
	"github.com/go-resty/resty/v2"
	"log"
	"strings"
	"time"
)

type Collector struct {
	metrics map[string]entities.Metric
	logger  *log.Logger
	client  *resty.Client
	hashKey string
}

func NewCollector(addr string, hashKey string, logger *log.Logger) Collector {
	createClient(addr)
	return Collector{
		client: createClient(addr).
			SetRetryWaitTime(1 * time.Second).
			SetRetryMaxWaitTime(5 * time.Second).
			SetRetryCount(3),
		metrics: make(map[string]entities.Metric),
		logger:  logger,
		hashKey: hashKey,
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

	dataJSON, err := json.Marshal(data)
	if err != nil {
		c.logger.Println("marshaling failed")
		return
	}

	compressedJSON, err := compress(dataJSON)
	if err != nil {
		c.logger.Println("compression failed")
		return
	}

	request := c.client.R()
	if c.hashKey != "" {
		hash := hashing.ComputeBase64URLHash(c.hashKey, dataJSON)
		request = request.SetHeader(hashing.HeaderHash, hash)
	}
	_, err = request.
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(compressedJSON).
		Post("/updates/")

	if err != nil {
		c.logger.Println("reporting metrics failed")
		return
	}

	c.logger.Println("metrics reported")
}

func compress(dataJSON []byte) ([]byte, error) {

	buf := bytes.Buffer{}
	cw := gzip.NewWriter(&buf)

	_, err := cw.Write(dataJSON)
	if err != nil {
		return nil, err
	}
	err = cw.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
