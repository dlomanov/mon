package binding

import (
	"errors"
	"github.com/dlomanov/mon/internal/handlers/apperrors"
	"github.com/dlomanov/mon/internal/handlers/metrics"
	"github.com/dlomanov/mon/internal/handlers/metrics/counter"
	"github.com/dlomanov/mon/internal/handlers/metrics/gauge"
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

	metricType, ok := metrics.ParseMetricType(raw.metricType)
	if !ok {
		err = ErrInvalidMetricType.New("unknown metric type %s", raw.metricType)
		return
	}
	if raw.name == "" {
		err = ErrInvalidMetricName.New("empty metric name")
		return
	}

	if raw.value == "" {
		err = ErrInvalidMetricValue.New("empty value")
		return
	}

	metric, e := createMetric(metricType, raw.name, raw.value)
	if e == nil {
		return
	}

	var appError apperrors.AppError
	if errors.As(e, &appError) {
		err = appError
		return
	}

	err = ErrInvalidMetricValue.New("invalid value type for %s metric", metricType)
	return
}

func createMetric(
	metricType metrics.MetricType,
	metricName string,
	metricValue string,
) (metric metrics.Metric, err error) {
	switch metricType {
	case metrics.MetricGauge:
		metric, err = gauge.NewMetric(metricName, metricValue)
	case metrics.MetricCounter:
		metric, err = counter.NewMetric(metricName, metricValue)
	default:
		err = ErrUnsupportedMetricType.New("unsupported %s metric")
	}
	return
}
