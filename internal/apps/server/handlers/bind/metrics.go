package bind

import (
	"encoding/json"
	"errors"
	"github.com/dlomanov/mon/internal/apperrors"
	"github.com/dlomanov/mon/internal/apps/apimodels"
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

func FromRouteParams(r *http.Request) (model apimodels.Metric, err error) {
	model.Name = chi.URLParam(r, "name")
	model.Type = chi.URLParam(r, "type")
	valueString := chi.URLParam(r, "value")

	metricType, ok := entities.ParseMetricType(model.Type)
	if !ok {
		return model, err
	}

	if metricType == entities.MetricGauge {
		var value float64
		value, err = strconv.ParseFloat(valueString, 64)
		if err != nil {
			err = errors.Join(ErrInvalidMetricValue.New(), err)
		} else {
			model.Value = &value
		}
	} else if metricType == entities.MetricCounter {
		var delta int64
		delta, err = strconv.ParseInt(valueString, 10, 64)
		if err != nil {
			err = errors.Join(ErrInvalidMetricValue.New(), err)
		} else {
			model.Delta = &delta
		}
	} else {
		err = ErrUnsupportedMetricType.New(model.Type)
	}

	return model, err
}

func FromJSON(r *http.Request) (model apimodels.Metric, err error) {
	if h := r.Header.Get("Content-Type"); !strings.HasPrefix(h, "application/json") {
		return model, ErrUnsupportedContentType.New(h)
	}
	err = json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		err = errors.Join(ErrInvalidMetricRequest.New(), err)
	}

	return model, err
}
