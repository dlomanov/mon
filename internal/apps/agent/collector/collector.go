package collector

import (
	"github.com/dlomanov/mon/internal/entities"
	"go.uber.org/zap"
)

type Collector struct {
	Metrics map[string]entities.Metric
	logger  *zap.Logger
}

func NewCollector(logger *zap.Logger) Collector {
	return Collector{
		Metrics: make(map[string]entities.Metric),
		logger:  logger,
	}
}
func (c *Collector) UpdateGauge(name string, value float64) {
	key := entities.MetricsKey{Name: name, Type: entities.MetricGauge}
	v := entities.Metric{MetricsKey: key, Value: &value}
	c.Metrics[key.String()] = v
}

func (c *Collector) UpdateCounter(name string, value int64) {
	key := entities.MetricsKey{Name: name, Type: entities.MetricCounter}
	keyString := key.String()
	v := entities.Metric{MetricsKey: key, Delta: &value}

	old, ok := c.Metrics[keyString]
	if ok {
		*v.Delta += *old.Delta
	}

	c.Metrics[keyString] = v
}

func (c *Collector) LogUpdated() {
	c.logger.Info("Metrics updated\n", zap.Int("updated_metric_count", len(c.Metrics)))
}
