package entities

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/dlomanov/mon/internal/apperrors"
)

type Metric struct {
	MetricsKey
	Value *float64
	Delta *int64
}

func NewMetric(key MetricsKey, value string) (Metric, error) {
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

type MetricsKey struct {
	Name string
	Type MetricType
}

func NewMetricsKey(value string) (metricsKey MetricsKey, err error) {
	values := strings.Split(value, "_")
	if len(values) < 2 {
		return metricsKey, errors.New("string value should contains separator '_'")
	}

	mtype, ok := ParseMetricType(values[0])
	if !ok {
		return metricsKey, fmt.Errorf("uknown metric type: %s", values[0])
	}

	return MetricsKey{Type: mtype, Name: values[1]}, nil
}

func (m *MetricsKey) String() string {
	return fmt.Sprintf("%s_%s", m.Type, m.Name)
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
	key := MetricsKey{Name: m.Name, Type: m.Type}
	return NewMetric(key, value)
}
