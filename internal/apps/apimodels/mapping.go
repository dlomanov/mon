package apimodels

import (
	"github.com/dlomanov/mon/internal/apperrors"
	"github.com/dlomanov/mon/internal/entities"
)

const (
	ErrUnsupportedMetricType = apperrors.ErrUnsupportedMetricType
	ErrInvalidMetricPath     = apperrors.ErrInvalidMetricPath
	ErrInvalidMetricType     = apperrors.ErrInvalidMetricType
	ErrInvalidMetricName     = apperrors.ErrInvalidMetricName
	ErrInvalidMetricValue    = apperrors.ErrInvalidMetricValue
)

func MapToEntity(model Metric) (entity entities.Metric, err error) {
	key, err := MapToEntityKey(model.MetricKey)
	if err != nil {
		return entity, err
	}

	entity.MetricsKey = key

	if key.Type == entities.MetricGauge && model.Value != nil {
		entity.Value = model.Value
	} else if key.Type == entities.MetricCounter && model.Delta != nil {
		entity.Delta = model.Delta
	} else {
		err = ErrInvalidMetricValue.Newf("invalid value for metrics type %s", key.Type)
	}

	return entity, err
}

func MapToEntityKey(key MetricKey) (entityKey entities.MetricsKey, err error) {
	metricType, ok := entities.ParseMetricType(key.Type)
	if !ok {
		err = ErrInvalidMetricType.Newf("unknown entity type %s", key.Type)
		return entityKey, err
	}

	if key.Id == "" {
		err = ErrInvalidMetricName.Newf("empty entity name")
		return entityKey, err
	}

	entityKey = entities.MetricsKey{
		Id:   key.Id,
		Type: metricType,
	}

	return entityKey, err
}

func MapToModel(entity entities.Metric) Metric {
	return Metric{
		MetricKey: MapToModelKey(entity.MetricsKey),
		Delta:     entity.Delta,
		Value:     entity.Value,
	}
}

func MapToModelKey(entity entities.MetricsKey) MetricKey {
	return MetricKey{
		Id:   entity.Id,
		Type: string(entity.Type),
	}
}
