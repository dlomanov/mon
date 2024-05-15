// Package bind provides functions to bind request parameters to metric models.
// It includes methods for extracting metric data from URL parameters and JSON request bodies.
package bind

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dlomanov/mon/internal/entities/apperrors"
	"net/http"
	"strconv"
	"strings"

	"github.com/dlomanov/mon/internal/apps/shared/apimodels"
	"github.com/dlomanov/mon/internal/entities"
	"github.com/go-chi/chi/v5"
)

var (
	ErrUnsupportedMetricType  = apperrors.ErrUnsupportedMetricType
	ErrUnsupportedContentType = apperrors.NewInvalid("unsupported content type")
	ErrInvalidMetricRequest   = apperrors.NewInvalid("invalid metric request")
	ErrInvalidMetricValue     = apperrors.NewInvalid("invalid metric value")
)

// MetricFromRouteParams binds metric data from URL parameters to a Metric model.
// It parses the metric type and value from the URL and returns a Metric model.
func MetricFromRouteParams(r *http.Request) (model apimodels.Metric, err error) {
	model.Name = chi.URLParam(r, "name")
	model.Type = chi.URLParam(r, "type")
	valueString := chi.URLParam(r, "value")

	metricType, ok := entities.ParseMetricType(model.Type)
	if !ok {
		return model, err
	}

	switch {
	case metricType == entities.MetricGauge:
		var value float64
		value, err = strconv.ParseFloat(valueString, 64)
		if err != nil {
			return model, errors.Join(ErrInvalidMetricValue, err)
		}
		model.Value = &value
	case metricType == entities.MetricCounter:
		var delta int64
		delta, err = strconv.ParseInt(valueString, 10, 64)
		if err != nil {
			return model, errors.Join(ErrInvalidMetricValue, err)
		}
		model.Delta = &delta
	default:
		return model, fmt.Errorf("%w: %s", ErrUnsupportedMetricType, model.Type)
	}

	return model, nil
}

// MetricFromJSON binds metric data from a JSON request body to a Metric model.
// It decodes the JSON body into a Metric model and returns it.
func MetricFromJSON(r *http.Request) (model apimodels.Metric, err error) {
	if h := r.Header.Get("Content-Type"); !strings.HasPrefix(h, "application/json") {
		return model, fmt.Errorf("%w: %s", ErrUnsupportedContentType, h)
	}
	err = json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		err = errors.Join(ErrInvalidMetricRequest, err)
	}

	return model, err
}

// MetricsFromJSON binds multiple metric data from a JSON request body to a slice of Metric models.
// It decodes the JSON body into a slice of Metric models and returns it.
func MetricsFromJSON(r *http.Request) (models []apimodels.Metric, err error) {
	if h := r.Header.Get("Content-Type"); !strings.HasPrefix(h, "application/json") {
		return models, fmt.Errorf("%w: %s", ErrUnsupportedContentType, h)
	}
	err = json.NewDecoder(r.Body).Decode(&models)
	if err != nil {
		err = errors.Join(ErrInvalidMetricRequest, err)
	}

	return models, err
}
