package apimodels

import (
	"github.com/dlomanov/mon/internal/apperrors"
	"github.com/dlomanov/mon/internal/entities"
)

const (
	ErrUnsupportedMetricType = apperrors.ErrUnsupportedMetricType

	ErrInvalidMetricType  = apperrors.ErrInvalidMetricType
	ErrInvalidMetricName  = apperrors.ErrInvalidMetricName
	ErrInvalidMetricValue = apperrors.ErrInvalidMetricValue
)

func MapToEntity(model Metric) (entity entities.Metric, err error) {
	key, err := MapToEntityKey(model.MetricKey)
	if err != nil {
		return entity, err
	}

	entity.MetricsKey = key
	switch {
	case key.Type == entities.MetricGauge && model.Value != nil:
		entity.Value = model.Value
	case key.Type == entities.MetricCounter && model.Delta != nil:
		entity.Delta = model.Delta
	default:
		err = ErrInvalidMetricValue.Newf("invalid value for metrics type %s", key.Type)
	}

	return entity, err
}

func MapToEntityKey(key MetricKey) (entityKey entities.MetricsKey, err error) {
	metricType, ok := entities.ParseMetricType(key.Type)

	if !ok {
		return entityKey, ErrInvalidMetricType.Newf("unknown entity type %s", key.Type)
	}

	if key.Name == "" {
		return entityKey, ErrInvalidMetricName.Newf("empty entity name")
	}

	return entities.MetricsKey{
		Name: key.Name,
		Type: metricType,
	}, nil
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
		Name: entity.Name,
		Type: string(entity.Type),
	}
}
