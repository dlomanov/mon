package endpoints

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dlomanov/mon/internal/apps/server/container"
	"github.com/dlomanov/mon/internal/apps/server/entrypoints/http/v1/endpoints/bind"
	"github.com/dlomanov/mon/internal/apps/server/usecases"
	"github.com/dlomanov/mon/internal/apps/shared/apimodels"
	"github.com/dlomanov/mon/internal/entities/apperrors"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"html/template"
	"net/http"
	"slices"
	"strings"
)

// HeaderContentType is the HTTP header key for specifying the content type of the response.
const HeaderContentType = "Content-Type"

var reportTemplate = template.Must(template.New("report").Parse(`{{range $val := .}}<p>{{$val}}</p>{{end}}`))

type metricEndpoint struct {
	logger        *zap.Logger
	metricUseCase *usecases.MetricUseCase
}

func UseMetrics(r chi.Router, c *container.Container) {
	e := &metricEndpoint{
		logger:        c.Logger,
		metricUseCase: c.MetricUseCase,
	}
	r.Get("/value/{type}/{name}", e.getByParams())
	r.Post("/value/", e.getByJSON())
	r.Get("/", e.report())
	r.Post("/update/{type}/{name}/{value}", e.updateByParams())
	r.Post("/update/", e.updateByJSON())
	r.Post("/updates/", e.updatesByJSON())
}

// getByParams
// @Summary		Get metric by parameters
// @Description	Retrieves a metric by its name and type using URL parameters.
// @ID				get_metric_by_params
//
// @Produce		plain
// @Param			type	path		string	true	"Type of the metric"
// @Param			name	path		string	true	"Name of the metric"
//
// @Success		200		{object}	string	"Metric value"
// @Failure		404		{object}	string	"Metric not found"
// @Router			/value/{type}/{name} [get]
func (e *metricEndpoint) getByParams() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := apimodels.MetricKey{
			Name: chi.URLParam(r, "name"),
			Type: chi.URLParam(r, "type"),
		}

		entityKey, err := apimodels.MapToEntityKey(key)
		if err != nil {
			e.logger.Debug("invalid request body", zap.Error(err))
			http.NotFound(w, r)
			return
		}

		entity, err := e.metricUseCase.Get(r.Context(), entityKey)
		var errNotFound *apperrors.AppErrorNotFound
		switch {
		case errors.As(err, &errNotFound):
			http.NotFound(w, r)
		case err != nil:
			w.WriteHeader(http.StatusInternalServerError)
			e.logger.Error("get entity failed", zap.Error(err))
		default:
			if _, err = w.Write([]byte(entity.StringValue())); err != nil {
				e.logger.Error("error occurred during response writing", zap.Error(err))
			}
		}
	}
}

// @Summary		Get metric by JSON
// @Description	Retrieves a metric by its name and type using a JSON request body.
// @ID				get_metric_by_json
//
// @Accept			json
// @Produce		json
//
// @Param			request	body		apimodels.MetricKey	true	"Metric key"
//
// @Success		200		{object}	apimodels.Metric
// @Failure		404		{object}	string	"Metric not found"
// @Failure		415		{object}	string	"Unsupported Media Type"
//
// @Router			/value/ [post]
func (e *metricEndpoint) getByJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h := r.Header.Get(HeaderContentType); !strings.HasPrefix(h, "application/json") {
			e.logger.Debug("invalid content-type", zap.String(HeaderContentType, h))
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		var key apimodels.MetricKey
		err := json.NewDecoder(r.Body).Decode(&key)
		if err != nil {
			e.logger.Debug("cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		entityKey, err := apimodels.MapToEntityKey(key)
		if err != nil {
			e.logger.Debug("invalid request body", zap.Error(err))
			http.NotFound(w, r)
			return
		}

		entity, err := e.metricUseCase.Get(r.Context(), entityKey)
		var errNotFound *apperrors.AppErrorNotFound
		switch {
		case errors.As(err, &errNotFound):
			http.NotFound(w, r)
		case err != nil:
			w.WriteHeader(http.StatusInternalServerError)
			e.logger.Error("get entity failed", zap.Error(err))
		default:
			metrics := apimodels.MapToModel(entity)
			w.Header().Set("Content-Type", "application/json")
			if err = json.NewEncoder(w).Encode(metrics); err != nil {
				e.logger.Error("error occurred during response writing", zap.Error(err))
			}
		}
	}
}

// @Summary		Generate a report
// @Description	Retrieves all metrics and generates a report in HTML format.
// @ID				generate_report
//
// @Produce		html
//
// @Success		200	{object}	string	"Report generated successfully"
// @Failure		500	{object}	string	"Failed to generate report"
//
// @Router			/report [get]
func (e *metricEndpoint) report() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		values, err := e.metricUseCase.GetAll(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			e.logger.Error("get entities failed", zap.Error(err))
			return
		}

		result := make([]string, 0, len(values))
		for _, v := range values {
			str := fmt.Sprintf("%s: %s\n", v.String(), v.StringValue())
			result = append(result, str)
		}
		slices.Sort(result)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = reportTemplate.Execute(w, result)
		if err != nil {
			e.logger.Error("error occurred", zap.String("error", err.Error()))
			return
		}
	}
}

// @Summary		Update metric by parameters
// @Description	Updates a metric by its name and type using URL parameters.
// @ID				update_metric_by_params
//
// @Param			type	path		string	true	"Type of the metric"
// @Param			name	path		string	true	"Name of the metric"
// @Param			value	path		string	true	"Value of the metric"
//
// @Success		200		{object}	string	"Metric updated successfully"
// @Failure		400		{object}	string	"Invalid metric parameters"
//
// @Router			/update/{type}/{name}/{value} [post]
func (e *metricEndpoint) updateByParams() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metric, err := bind.MetricFromRouteParams(r)
		if err != nil {
			e.logger.Error("error occurred during model binding", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}
		entity, err := apimodels.MapToEntity(metric)
		if err != nil {
			e.logger.Error("error occurred during model mapping", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		_, err = e.metricUseCase.Update(r.Context(), entity)
		if err != nil {
			e.logger.Error("error occurred during metric update", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// @Summary		Update metrics by JSON
// @Description	Updates multiple metrics using a JSON request body.
// @ID				update_metrics_by_json
//
// @Accept			json
// @Produce		json
//
// @Param			metrics	body		[]apimodels.Metric	true	"Metrics to update"
//
// @Success		200		{object}	string				"Metrics updated successfully"
// @Failure		400		{object}	string				"Invalid metrics JSON"
//
// @Router			/updates/ [post]
func (e *metricEndpoint) updatesByJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics, err := bind.MetricsFromJSON(r)
		if err != nil {
			e.logger.Error("error occurred during model binding", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		values, err := apimodels.MapToEntities(metrics)
		if err != nil {
			e.logger.Error("error occurred during model mapping", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		_, err = e.metricUseCase.Update(r.Context(), values...)
		if err != nil {
			e.logger.Error("error occurred during metric update", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// @Summary		Update metric by JSON
// @Description	Updates a metric using a JSON request body.
// @ID				update_metric_by_json
//
// @Accept			json
// @Produce		json
//
// @Param			metric	body		apimodels.Metric	true	"Metric to update"
//
// @Success		200		{object}	apimodels.Metric	"Updated metric"
// @Failure		400		{object}	string				"Invalid metric JSON"
//
// @Router			/update/ [post]
func (e *metricEndpoint) updateByJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metric, err := bind.MetricFromJSON(r)
		if err != nil {
			e.logger.Error("error occurred during model binding", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}
		entity, err := apimodels.MapToEntity(metric)
		if err != nil {
			e.logger.Error("error occurred during model mapping", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}

		processed, err := e.metricUseCase.Update(r.Context(), entity)
		if err != nil {
			e.logger.Error("error occurred during metric update", zap.Error(err))
			w.WriteHeader(statusCode(err))
			return
		}
		w.Header().Set(HeaderContentType, "application/json")
		err = json.NewEncoder(w).Encode(apimodels.MapToModel(processed[0]))
		if err != nil {
			e.logger.Error("error occurred during response writing", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func statusCode(err error) int {
	switch {
	case errors.Is(err, bind.ErrUnsupportedContentType):
		return http.StatusUnsupportedMediaType
	case errors.Is(err, bind.ErrInvalidMetricRequest):
		return http.StatusBadRequest
	case errors.Is(err, apimodels.ErrInvalidMetricType):
		return http.StatusBadRequest
	case errors.Is(err, apimodels.ErrInvalidMetricName):
		return http.StatusNotFound
	case errors.Is(err, apimodels.ErrInvalidMetricValue):
		return http.StatusBadRequest
	case errors.Is(err, apimodels.ErrUnsupportedMetricType):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
