package apimodels

import (
	"fmt"
	"github.com/dlomanov/mon/internal/entities"
	"github.com/dlomanov/mon/internal/entities/apperrors"
)

var (
	ErrUnsupportedMetricType = apperrors.ErrUnsupportedMetricType
	ErrInvalidMetricType     = apperrors.NewInvalid("invalid metric type")
	ErrInvalidMetricName     = apperrors.NewInvalid("invalid metric name")
	ErrInvalidMetricValue    = apperrors.NewInvalid("invalid metric value")
)

func MapToEntities(models []Metric) (values []entities.Metric, err error) {
	values = make([]entities.Metric, 0, len(models))

	for _, v := range models {
		entity, err := MapToEntity(v)
		if err != nil {
			return nil, err
		}
		values = append(values, entity)
	}

	return values, nil
}

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
		err = fmt.Errorf("%w: %s", ErrInvalidMetricValue, key.Type)
	}

	return entity, err
}

func MapToEntityKey(key MetricKey) (entityKey entities.MetricsKey, err error) {
	metricType, ok := entities.ParseMetricType(key.Type)

	if !ok {
		return entityKey, fmt.Errorf("%w: %s", ErrUnsupportedMetricType, key.Type)
	}

	if key.Name == "" {
		return entityKey, ErrInvalidMetricName
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
