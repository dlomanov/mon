package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dlomanov/mon/internal/apperrors"
	"github.com/dlomanov/mon/internal/apps/server/container"
	"github.com/dlomanov/mon/internal/apps/server/handlers/bind"
	"github.com/dlomanov/mon/internal/apps/shared/apimodels"
	"github.com/dlomanov/mon/internal/entities"
	"go.uber.org/zap"
)

// @Summary Update metric by parameters
// @Description Updates a metric by its name and type using URL parameters.
// @ID update_metric_by_params
//
// @Param type path string true "Type of the metric"
// @Param name path string true "Name of the metric"
// @Param value path string true "Value of the metric"
//
// @Success 200 {object} string "Metric updated successfully"
// @Failure 400 {object} string "Invalid metric parameters"
//
// @Router /update/{type}/{name}/{value} [post]
func UpdateByParams(c *container.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metric, err := bind.MetricFromRouteParams(r)
		if err != nil {
			c.Logger.Error("error occurred during model binding", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		entity, err := apimodels.MapToEntity(metric)
		if err != nil {
			c.Logger.Error("error occurred during model mapping", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		_, err = handle(r.Context(), c.Storage, false, entity)
		if err != nil {
			c.Logger.Error("error occurred during metric update", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// @Summary Update metrics by JSON
// @Description Updates multiple metrics using a JSON request body.
// @ID update_metrics_by_json
//
// @Accept json
// @Produce json
//
// @Param metrics body []apimodels.Metric true "Metrics to update"
//
// @Success 200 {object} string "Metrics updated successfully"
// @Failure 400 {object} string "Invalid metrics JSON"
//
// @Router /updates/ [post]
func UpdatesByJSON(c *container.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics, err := bind.MetricsFromJSON(r)
		if err != nil {
			c.Logger.Error("error occurred during model binding", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		values, err := apimodels.MapToEntities(metrics)
		if err != nil {
			c.Logger.Error("error occurred during model mapping", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		_, err = handle(r.Context(), c.Storage, false, values...)
		if err != nil {
			c.Logger.Error("error occurred during metric update", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// @Summary Update metric by JSON
// @Description Updates a metric using a JSON request body.
// @ID update_metric_by_json
//
// @Accept json
// @Produce json
//
// @Param metric body apimodels.Metric true "Metric to update"
//
// @Success 200 {object} apimodels.Metric "Updated metric"
// @Failure 400 {object} string "Invalid metric JSON"
//
// @Router /update/ [post]
func UpdateByJSON(c *container.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metric, err := bind.MetricFromJSON(r)
		if err != nil {
			c.Logger.Error("error occurred during model binding", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		entity, err := apimodels.MapToEntity(metric)
		if err != nil {
			c.Logger.Error("error occurred during model mapping", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		processed, err := handle(r.Context(), c.Storage, true, entity)
		if err != nil {
			c.Logger.Error("error occurred during metric update", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		w.Header().Set(HeaderContentType, "application/json")
		err = json.NewEncoder(w).Encode(apimodels.MapToModel(processed[0]))
		if err != nil {
			c.Logger.Error("error occurred during response writing", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func handle(
	ctx context.Context,
	storage container.Storage,
	needResult bool,
	metrics ...entities.Metric,
) (processedMetrics []entities.Metric, err error) {
	processedMetrics = make([]entities.Metric, 0)

	for _, entity := range metrics {
		var processed entities.Metric
		switch entity.Type {
		case entities.MetricGauge:
			processed, err = HandleGauge(ctx, entity, storage)
		case entities.MetricCounter:
			processed, err = HandleCounter(ctx, entity, storage)
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

func HandleGauge(
	ctx context.Context,
	metric entities.Metric,
	storage container.Storage,
) (entities.Metric, error) {
	err := storage.Set(ctx, metric)
	return metric, err
}

func HandleCounter(
	ctx context.Context,
	metric entities.Metric,
	storage container.Storage,
) (result entities.Metric, err error) {
	old, ok, err := storage.Get(ctx, metric.MetricsKey)
	if err != nil {
		return result, err
	}

	if ok {
		*metric.Delta += *old.Delta
	}

	return metric, storage.Set(ctx, metric)
}
