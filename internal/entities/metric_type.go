package entities

import (
	"strings"

	"github.com/dlomanov/mon/internal/apperrors"
)

// MetricType represents the type of a metric.
type MetricType string

const (
	MetricGauge   MetricType = "gauge"
	MetricCounter MetricType = "counter"
)

func (t MetricType) IsValid() bool {
	return t == MetricGauge || t == MetricCounter
}

// MustParseMetricType attempts to parse a string into a MetricType.
// If the string does not match any known MetricType, it panics.
func MustParseMetricType(str string) MetricType {
	value, ok := ParseMetricType(str)
	if !ok {
		panic(apperrors.ErrUnsupportedMetricType.New(str))
	}
	return value
}

// ParseMetricType attempts to parse a string into a MetricType.
// It returns the parsed MetricType and a boolean indicating whether the parsing was successful.
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
