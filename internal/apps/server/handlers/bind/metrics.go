package bind

import (
	"encoding/json"
	"errors"
	"github.com/dlomanov/mon/internal/apperrors"
	"github.com/dlomanov/mon/internal/apps/shared/apimodels"
	"github.com/dlomanov/mon/internal/entities"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"strings"
)

const (
	ErrUnsupportedMetricType  = apperrors.ErrUnsupportedMetricType
	ErrUnsupportedContentType = apperrors.ErrUnsupportedContentType
	ErrInvalidMetricRequest   = apperrors.ErrInvalidMetricRequest
	ErrInvalidMetricValue     = apperrors.ErrInvalidMetricValue
)

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
			return model, errors.Join(ErrInvalidMetricValue.New(), err)
		}
		model.Value = &value
	case metricType == entities.MetricCounter:
		var delta int64
		delta, err = strconv.ParseInt(valueString, 10, 64)
		if err != nil {
			return model, errors.Join(ErrInvalidMetricValue.New(), err)
		}
		model.Delta = &delta
	default:
		return model, ErrUnsupportedMetricType.New(model.Type)
	}

	return model, nil
}

func MetricFromJSON(r *http.Request) (model apimodels.Metric, err error) {
	if h := r.Header.Get("Content-Type"); !strings.HasPrefix(h, "application/json") {
		return model, ErrUnsupportedContentType.New(h)
	}
	err = json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		err = errors.Join(ErrInvalidMetricRequest.New(), err)
	}

	return model, err
}

func MetricsFromJSON(r *http.Request) (models []apimodels.Metric, err error) {
	if h := r.Header.Get("Content-Type"); !strings.HasPrefix(h, "application/json") {
		return models, ErrUnsupportedContentType.New(h)
	}
	err = json.NewDecoder(r.Body).Decode(&models)
	if err != nil {
		err = errors.Join(ErrInvalidMetricRequest.New(), err)
	}

	return models, err
}
