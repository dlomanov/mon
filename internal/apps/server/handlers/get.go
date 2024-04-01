package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dlomanov/mon/internal/apps/server/container"
	"github.com/dlomanov/mon/internal/apps/shared/apimodels"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// @Summary Get metric by parameters
// @Description Retrieves a metric by its name and type using URL parameters.
// @ID get_metric_by_params
//
// @Produce text
// @Param type path string true "Type of the metric"
// @Param name path string true "Name of the metric"
//
// @Success 200 {object} string "Metric value"
// @Failure 404 {object} string "Metric not found"
// @Router /value/{type}/{name} [get]
func GetByParams(c *container.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := apimodels.MetricKey{
			Name: chi.URLParam(r, "name"),
			Type: chi.URLParam(r, "type"),
		}

		entityKey, err := apimodels.MapToEntityKey(key)
		if err != nil {
			c.Logger.Debug("invalid request body", zap.Error(err))
			http.NotFound(w, r)
			return
		}

		entity, ok, err := c.Storage.Get(r.Context(), entityKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			c.Logger.Error("get entity failed", zap.Error(err))
			return
		}
		if !ok {
			http.NotFound(w, r)
			return
		}

		_, err = w.Write([]byte(entity.StringValue()))
		if err != nil {
			c.Logger.Error("error occurred during response writing", zap.Error(err))
		}
	}
}

// @Summary Get metric by JSON
// @Description Retrieves a metric by its name and type using a JSON request body.
// @ID get_metric_by_json
//
// @Accept json
// @Produce json
//
// @Param MetricKey body apimodels.MetricKey true "Metric key"
//
// @Success 200 {object} apimodels.Metric
// @Failure 404 {object} string "Metric not found"
// @Failure 415 {object} string "Unsupported Media Type"
//
// @Router /value/ [post]
func GetByJSON(c *container.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h := r.Header.Get(HeaderContentType); !strings.HasPrefix(h, "application/json") {
			c.Logger.Debug("invalid content-type", zap.String(HeaderContentType, h))
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		var key apimodels.MetricKey
		err := json.NewDecoder(r.Body).Decode(&key)
		if err != nil {
			c.Logger.Debug("cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		entityKey, err := apimodels.MapToEntityKey(key)
		if err != nil {
			c.Logger.Debug("invalid request body", zap.Error(err))
			http.NotFound(w, r)
			return
		}

		entity, ok, err := c.Storage.Get(r.Context(), entityKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			c.Logger.Error("get entity failed", zap.Error(err))
			return
		}
		if !ok {
			http.NotFound(w, r)
			return
		}

		metrics := apimodels.MapToModel(entity)
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(metrics)
		if err != nil {
			c.Logger.Error("error occurred", zap.Error(err))
		}
	}
}
