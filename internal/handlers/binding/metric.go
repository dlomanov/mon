package binding

import (
	"errors"
	"github.com/dlomanov/mon/internal/handlers/apperrors"
	"github.com/dlomanov/mon/internal/handlers/metrics"
	"strconv"
	"strings"
)

const (
	fieldCount = 3
	sep        = "/"
)

const (
	ErrInvalidMetricPath     apperrors.Code = "ERR_VALIDATION_INVALID_METRIC_PATH"
	ErrInvalidMetricType     apperrors.Code = "ERR_VALIDATION_INVALID_METRIC_TYPE"
	ErrInvalidMetricName     apperrors.Code = "ERR_VALIDATION_INVALID_METRIC_NAME"
	ErrInvalidMetricValue    apperrors.Code = "ERR_VALIDATION_INVALID_METRIC_VALUE"
	ErrUnsupportedMetricType apperrors.Code = "ERR_INTERNAL_UNSUPPORTED_METRIC_TYPE"
)

func Metric(path string) (metric metrics.Metric, err error) {
	trimmed := strings.TrimRight(path, sep)
	values := strings.Split(trimmed, sep)

	raw := struct {
		metricType string
		name       string
		value      string
	}{}

	for i, v := range values {
		switch i {
		case 0:
			raw.metricType = strings.ToLower(v)
		case 1:
			raw.name = strings.ToLower(v)
		case 2:
			raw.value = v
		default:
			err = ErrInvalidMetricPath.New("expected %d path values, but received %d", fieldCount, len(values))
		}
	}

	t, ok := metrics.ParseMetricType(raw.metricType)
	if !ok {
		err = ErrInvalidMetricType.New("unknown metric type %s", raw.metricType)
		return
	}
	metric.Type = t

	if raw.name == "" {
		err = ErrInvalidMetricName.New("empty metric name")
		return
	}
	metric.Name = raw.name

	if raw.value == "" {
		err = ErrInvalidMetricValue.New("empty value")
		return
	}

	v, e := parseValue(metric.Type, raw.value)
	if e == nil {
		metric.Value = v
		return
	}

	var appError apperrors.AppError
	if errors.As(e, &appError) {
		err = appError
		return
	}

	err = ErrInvalidMetricValue.New("invalid value type for %s metric", metric.Type)
	return
}

func parseValue(t metrics.MetricType, rawValue string) (value any, err error) {
	switch t {
	case metrics.MetricGauge:
		value, err = strconv.ParseFloat(rawValue, 64)
	case metrics.MetricCounter:
		value, err = strconv.ParseInt(rawValue, 10, 64)
	default:
		err = ErrUnsupportedMetricType.New("unsupported %s metric", t)
	}
	return
}
