package entities

import (
	"strings"

	"github.com/dlomanov/mon/internal/apperrors"
)

type MetricType string

const (
	MetricGauge   MetricType = "gauge"
	MetricCounter MetricType = "counter"
)

func (t MetricType) IsValid() bool {
	return t == MetricGauge || t == MetricCounter
}

func ParseMetricTypeForced(str string) MetricType {
	value, ok := ParseMetricType(str)
	if !ok {
		panic(apperrors.ErrUnsupportedMetricType.New(str))
	}
	return value
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
