package metrics

import (
	"fmt"
	"strings"
)

type MetricType string

const (
	MetricGauge   MetricType = "gauge"
	MetricCounter MetricType = "counter"
)

func (t MetricType) IsValid() bool {
	return t == MetricGauge || t == MetricCounter
}

func ParseMetricType(str string) (value MetricType, ok bool) {
	switch lower := strings.ToLower(str); lower {
	case string(MetricGauge):
		return MetricGauge, true
	case string(MetricCounter):
		return MetricCounter, true
	default:
		return "", false
	}
}

func (t MetricType) CreateKey(metricName string) string {
	return fmt.Sprintf("%s_%s", t, metricName)
}
