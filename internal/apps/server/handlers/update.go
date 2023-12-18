package handlers

import (
	"errors"
	"github.com/dlomanov/mon/internal/apperrors"
	bind "github.com/dlomanov/mon/internal/apps/server/handlers/binding"
	"github.com/dlomanov/mon/internal/entities/metrics/counter"
	"github.com/dlomanov/mon/internal/entities/metrics/gauge"
	"github.com/dlomanov/mon/internal/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func Update(db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var apperr apperrors.AppError

		metric, err := bind.Metric(bind.RawMetric{
			Type:  chi.URLParam(r, "type"),
			Name:  chi.URLParam(r, "name"),
			Value: chi.URLParam(r, "value"),
		})
		if errors.As(err, &apperr) {
			w.WriteHeader(statusCode(apperr))
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = nil
		switch v := metric.(type) {
		case gauge.Metric:
			err = HandleGauge(v, db)
		case counter.Metric:
			err = HandleCounter(v, db)
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
	case bind.ErrInvalidMetricPath:
		return http.StatusNotFound
	case bind.ErrInvalidMetricType:
		return http.StatusBadRequest
	case bind.ErrInvalidMetricName:
		return http.StatusNotFound
	case bind.ErrInvalidMetricValue:
		return http.StatusBadRequest
	case bind.ErrUnsupportedMetricType:
		return http.StatusInternalServerError
	case ErrInvalidMetricValueType:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

const (
	ErrInvalidMetricValueType apperrors.Code = "ERR_VALIDATION_INVALID_METRIC_VALUE_TYPE"
)

func HandleGauge(metric gauge.Metric, db storage.Storage) error {
	db.Set(metric.Key(), metric.StringValue())
	return nil
}

func HandleCounter(metric counter.Metric, db storage.Storage) (err error) {
	value, ok := db.Get(metric.Key())
	if !ok {
		db.Set(metric.Key(), metric.StringValue())
		return
	}

	old, err := metric.With(value)
	if err != nil {
		return
	}

	metric.Value = metric.Value + old.Value
	db.Set(metric.Key(), metric.StringValue())
	return
}
