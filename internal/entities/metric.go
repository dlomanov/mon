package entities

import (
	"fmt"
	"github.com/dlomanov/mon/internal/apperrors"
	"strconv"
)

type Metric struct {
	MetricsKey
	Value *float64
	Delta *int64
}

type MetricsKey struct {
	ID   string
	Type MetricType
}

func (m *MetricsKey) String() string {
	return fmt.Sprintf("%s_%s", m.Type, m.ID)
}

func (m *Metric) StringValue() string {
	switch m.Type {
	case MetricCounter:
		return strconv.FormatInt(*m.Delta, 10)
	case MetricGauge:
		return strconv.FormatFloat(*m.Value, 'f', -1, 64)
	default:
		panic(fmt.Sprintf("unsupported metric type %s", m.Type))
	}
}

func (m *Metric) CloneWith(value string) (Metric, error) {
	key := MetricsKey{ID: m.ID, Type: m.Type}

	if key.Type == MetricGauge {
		v, err := strconv.ParseFloat(value, 64)
		return Metric{
			MetricsKey: key,
			Value:      &v,
		}, err
	}

	if key.Type == MetricCounter {
		v, err := strconv.ParseInt(value, 10, 64)
		return Metric{
			MetricsKey: key,
			Delta:      &v,
		}, err

	}

	return Metric{}, apperrors.ErrUnsupportedMetricType.New(key.Type)
}
