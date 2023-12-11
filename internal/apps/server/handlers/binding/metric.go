package binding

import (
	"errors"
	"github.com/dlomanov/mon/internal/apperrors"
	"github.com/dlomanov/mon/internal/entities/metrics"
	"github.com/dlomanov/mon/internal/entities/metrics/counter"
	"github.com/dlomanov/mon/internal/entities/metrics/gauge"
)

const (
	ErrInvalidMetricPath     apperrors.Code = "ERR_VALIDATION_INVALID_METRIC_PATH"
	ErrInvalidMetricType     apperrors.Code = "ERR_VALIDATION_INVALID_METRIC_TYPE"
	ErrInvalidMetricName     apperrors.Code = "ERR_VALIDATION_INVALID_METRIC_NAME"
	ErrInvalidMetricValue    apperrors.Code = "ERR_VALIDATION_INVALID_METRIC_VALUE"
	ErrUnsupportedMetricType apperrors.Code = "ERR_INTERNAL_UNSUPPORTED_METRIC_TYPE"
)

type RawMetric struct {
	Type  string
	Name  string
	Value string
}

func Metric(raw RawMetric) (metric metrics.Metric, err error) {
	metricType, ok := metrics.ParseMetricType(raw.Type)
	if !ok {
		err = ErrInvalidMetricType.New("unknown metric type %s", raw.Type)
		return
	}
	if raw.Name == "" {
		err = ErrInvalidMetricName.New("empty metric raw.Name")
		return
	}

	if raw.Value == "" {
		err = ErrInvalidMetricValue.New("empty value")
		return
	}

	metric, e := createMetric(metricType, raw.Name, raw.Value)
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
