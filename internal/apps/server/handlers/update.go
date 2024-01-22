package handlers

import (
	"encoding/json"
	"errors"
	"github.com/dlomanov/mon/internal/apperrors"
	"github.com/dlomanov/mon/internal/apps/apimodels"
	"github.com/dlomanov/mon/internal/apps/server/handlers/bind"
	"github.com/dlomanov/mon/internal/apps/server/handlers/interfaces"
	"github.com/dlomanov/mon/internal/entities"
	"go.uber.org/zap"
	"net/http"
)

func UpdateByParams(logger *zap.Logger, db interfaces.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metric, err := bind.MetricFromRouteParams(r)
		if err != nil {
			logger.Error("error occurred during model binding", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		entity, err := apimodels.MapToEntity(metric)
		if err != nil {
			logger.Error("error occurred during model mapping", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		_, err = handle(db, false, entity)
		if err != nil {
			logger.Error("error occurred during metric update", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func UpdatesByJSON(logger *zap.Logger, db interfaces.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics, err := bind.MetricsFromJSON(r)
		if err != nil {
			logger.Error("error occurred during model binding", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		values, err := apimodels.MapToEntities(metrics)
		if err != nil {
			logger.Error("error occurred during model mapping", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		_, err = handle(db, false, values...)
		if err != nil {
			logger.Error("error occurred during metric update", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func UpdateByJSON(logger *zap.Logger, db interfaces.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metric, err := bind.MetricFromJSON(r)
		if err != nil {
			logger.Error("error occurred during model binding", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		entity, err := apimodels.MapToEntity(metric)
		if err != nil {
			logger.Error("error occurred during model mapping", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		processed, err := handle(db, true, entity)
		if err != nil {
			logger.Error("error occurred during metric update", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		w.Header().Set(HeaderContentType, "application/json")
		err = json.NewEncoder(w).Encode(apimodels.MapToModel(processed[0]))
		if err != nil {
			logger.Error("error occurred during response writing", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func handle(
	db interfaces.Storage,
	needResult bool,
	metrics ...entities.Metric,
) (processedMetrics []entities.Metric, err error) {
	processedMetrics = make([]entities.Metric, 0)

	for _, entity := range metrics {
		var processed entities.Metric
		switch entity.Type {
		case entities.MetricGauge:
			processed, err = HandleGauge(entity, db)
		case entities.MetricCounter:
			processed, err = HandleCounter(entity, db)
		default:
			err = apperrors.ErrUnsupportedMetricType.New(entity.Type)
		}
		if err != nil {
			return nil, err
		}

		if needResult {
			processedMetrics = append(processedMetrics, processed)
		}
	}

	return processedMetrics, nil
}

func statusCode(err error) int {
	var apperr apperrors.AppError
	if !errors.As(err, &apperr) {
		return http.StatusInternalServerError
	}

	switch apperr.Type {
	case bind.ErrUnsupportedContentType:
		return http.StatusUnsupportedMediaType
	case bind.ErrInvalidMetricRequest:
		return http.StatusBadRequest
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

func HandleGauge(metric entities.Metric, db interfaces.Storage) (entities.Metric, error) {
	db.Set(metric)
	return metric, nil
}

func HandleCounter(metric entities.Metric, db interfaces.Storage) (entities.Metric, error) {
	key := metric.MetricsKey

	old, ok := db.Get(key)
	if !ok {
		db.Set(metric)
		return metric, nil
	}

	*metric.Delta += *old.Delta
	db.Set(metric)
	return metric, nil
}
