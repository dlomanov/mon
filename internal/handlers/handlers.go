package handlers

import (
	"errors"
	"github.com/dlomanov/mon/internal/handlers/apperrors"
	bind "github.com/dlomanov/mon/internal/handlers/binding"
	metrics "github.com/dlomanov/mon/internal/handlers/metrics"
	"github.com/dlomanov/mon/internal/storage"
	"net/http"
	"strconv"
	"strings"
)

// UpdateHandler path: /update/<type>/<name>/<value>
func UpdateHandler(path string, db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var apperr apperrors.AppError

		pathValues := strings.TrimLeft(r.RequestURI, path)
		metric, err := bind.Metric(pathValues)
		if errors.As(err, &apperr) {
			w.WriteHeader(statusCode(apperr))
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = nil
		switch metric.Type {
		case metrics.MetricGauge:
			err = HandleGauge(metric, db)
		case metrics.MetricCounter:
			err = HandleCounter(metric, db)
		default:
			w.WriteHeader(http.StatusNotImplemented)
			return
		}
		if err == nil {
			w.WriteHeader(http.StatusOK)
			return
		}

		if errors.As(err, &apperr) {
			w.WriteHeader(statusCode(apperr))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func statusCode(a apperrors.AppError) int {
	switch a.Code {
	case bind.InvalidMetricPath:
		return http.StatusNotFound
	case bind.InvalidMetricType:
		return http.StatusBadRequest
	case bind.InvalidMetricName:
		return http.StatusNotFound
	case bind.InvalidMetricValue:
		return http.StatusBadRequest
	case bind.UnsupportedMetricType:
		return http.StatusInternalServerError
	case InvalidMetricValueType:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

const (
	InvalidMetricValueType apperrors.Code = "ERR_VALIDATION_INVALID_METRIC_VALUE_TYPE"
)

func HandleGauge(metric metrics.Metric, db storage.Storage) error {
	v, ok := metric.Value.(float64)
	if !ok {
		return InvalidMetricValueType.New("invalid value type for %s metric", metric.Type)
	}

	valueString := strconv.FormatFloat(v, 'f', -1, 64)
	db.Set(metric.Key(), valueString)
	return nil
}

func HandleCounter(metric metrics.Metric, db storage.Storage) error {
	v, ok := metric.Value.(int64)
	if !ok {
		return InvalidMetricValueType.New("invalid value type for %s metric", metric.Type)
	}

	oldString, ok := db.Get(metric.Key())
	if !ok {
		db.Set(metric.Key(), strconv.FormatInt(v, 10))
		return nil
	}

	old, err := strconv.ParseInt(oldString, 10, 64)
	if err != nil {
		return err
	}

	newValue := v + old
	db.Set(metric.Key(), strconv.FormatInt(newValue, 10))
	return nil
}
