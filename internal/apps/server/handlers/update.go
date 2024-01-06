package handlers

import (
	"encoding/json"
	"errors"
	"github.com/dlomanov/mon/internal/apperrors"
	"github.com/dlomanov/mon/internal/apps/apimodels"
	"github.com/dlomanov/mon/internal/apps/server/logger"
	"github.com/dlomanov/mon/internal/entities"
	"github.com/dlomanov/mon/internal/storage"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

const (
	contentTypeKey = "Content-Type"
)

func Update(db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var metrics apimodels.Metric

		if h := r.Header.Get(contentTypeKey); !strings.HasPrefix(h, "application/json") {
			logger.Log.Debug("invalid content-type", zap.String(contentTypeKey, h))
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		if err = json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		entity, err := apimodels.MapToEntity(metrics)
		if err != nil {
			logger.Log.Error("error occurred during model mapping", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		processed, err := handle(entity, db)
		if err != nil {
			logger.Log.Error("error occurred during metric update", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		w.Header().Set(contentTypeKey, "application/json")
		err = json.NewEncoder(w).Encode(apimodels.MapToModel(processed))
		if err != nil {
			logger.Log.Error("error occurred during response writing", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func handle(entity entities.Metric, db storage.Storage) (processed entities.Metric, err error) {
	switch entity.Type {
	case entities.MetricGauge:
		processed, err = HandleGauge(entity, db)
	case entities.MetricCounter:
		processed, err = HandleCounter(entity, db)
	default:
		err = apperrors.ErrUnsupportedMetricType.New(entity.Type)
	}

	return processed, err
}

func statusCode(err error) int {
	var apperr apperrors.AppError
	if !errors.As(err, &apperr) {
		return http.StatusInternalServerError
	}

	switch apperr.Type {
	case apimodels.ErrInvalidMetricPath:
		return http.StatusNotFound
	case apimodels.ErrInvalidMetricType:
		return http.StatusBadRequest
	case apimodels.ErrInvalidMetricName:
		return http.StatusNotFound
	case apimodels.ErrInvalidMetricValue:
		return http.StatusBadRequest
	case apimodels.ErrUnsupportedMetricType:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func HandleGauge(metric entities.Metric, db storage.Storage) (entities.Metric, error) {
	db.Set(metric.MetricsKey.String(), metric.StringValue())
	return metric, nil
}

func HandleCounter(metric entities.Metric, db storage.Storage) (entities.Metric, error) {
	key := metric.MetricsKey.String()

	value, ok := db.Get(key)
	if !ok {
		db.Set(key, metric.StringValue())
		return metric, nil
	}

	old, err := metric.CloneWith(value)
	if err != nil {
		return entities.Metric{}, err
	}

	*metric.Delta += *old.Delta
	db.Set(key, metric.StringValue())
	return metric, nil
}
