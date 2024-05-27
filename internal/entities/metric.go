package entities

import (
	"errors"
	"fmt"
	"github.com/dlomanov/mon/internal/entities/apperrors"
	"strconv"
	"strings"
)

// Metric represents a metric with a key and optional delta or value.
type Metric struct {
	MetricsKey
	Value *float64
	Delta *int64
}

// NewMetric creates a new Metric instance based on the provided key and value string.
// It parses the value string into either a float64 for gauge metrics or an int64 for counter metrics.
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

	return Metric{}, fmt.Errorf("%w: %s", apperrors.ErrUnsupportedMetricType, key.Type)
}

// MetricsKey is a unique identifier for a metric, consisting of a name and type.
type MetricsKey struct {
	Name string
	Type MetricType
}

// NewMetricsKey parses a string into a MetricsKey, which includes the metric type and name.
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

// String returns a string representation of the MetricsKey, formatted as "type_name".
func (m *MetricsKey) String() string {
	return fmt.Sprintf("%s_%s", m.Type, m.Name)
}

// StringValue returns the string representation of the metric's value, formatted according to its type.
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

// CloneWith creates a new Metric with the same key but a different value.
func (m *Metric) CloneWith(value string) (Metric, error) {
	key := MetricsKey{Name: m.Name, Type: m.Type}
	return NewMetric(key, value)
}
