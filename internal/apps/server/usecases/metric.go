package usecases

import (
	"context"
	"fmt"
	"github.com/dlomanov/mon/internal/entities"
	"github.com/dlomanov/mon/internal/entities/apperrors"
)

type (
	MetricUseCase struct {
		storage Storage
	}

	Storage interface {
		Set(ctx context.Context, metrics ...entities.Metric) error
		Get(ctx context.Context, key entities.MetricsKey) (metric entities.Metric, ok bool, err error)
		All(ctx context.Context) (result []entities.Metric, err error)
	}
)

func NewMetricUseCase(storage Storage) *MetricUseCase {
	return &MetricUseCase{
		storage: storage,
	}
}

func (uc *MetricUseCase) Get(ctx context.Context, key entities.MetricsKey) (entities.Metric, error) {
	m, ok, err := uc.storage.Get(ctx, key)
	switch {
	case err != nil:
		return entities.Metric{}, fmt.Errorf("%w: %w", apperrors.NewInternal("failed to get metric"), err)
	case !ok:
		return entities.Metric{}, apperrors.NewNotFound("metric not found")
	default:
		return m, nil
	}
}

func (uc *MetricUseCase) GetAll(ctx context.Context) ([]entities.Metric, error) {
	m, err := uc.storage.All(ctx)
	switch {
	case err != nil:
		return nil, fmt.Errorf("%w: %w", apperrors.NewInternal("failed to get metric"), err)
	default:
		return m, nil
	}
}

func (uc *MetricUseCase) Update(
	ctx context.Context,
	metrics ...entities.Metric,
) ([]entities.Metric, error) {
	result := make([]entities.Metric, 0, len(metrics))
	for _, metric := range metrics {
		m, err := uc.update(ctx, metric)
		if err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, nil
}

func (uc *MetricUseCase) update(ctx context.Context, metric entities.Metric) (entities.Metric, error) {
	switch metric.MetricsKey.Type {
	case entities.MetricGauge:
		return metric, uc.storage.Set(ctx, metric)
	case entities.MetricCounter:
		old, ok, err := uc.storage.Get(ctx, metric.MetricsKey)
		if err != nil {
			return metric, fmt.Errorf("%w: %w", apperrors.NewInternal("failed to update metric"), err)
		}
		if ok {
			*metric.Delta += *old.Delta
		}
		return metric, uc.storage.Set(ctx, metric)
	default:
		return metric, apperrors.ErrUnsupportedMetricType
	}
}
