package storage_test

import (
	"context"
	"github.com/dlomanov/mon/internal/infra/storage"
	"testing"

	"github.com/dlomanov/mon/internal/entities"
	"github.com/stretchr/testify/require"
)

func TestMemStorage(t *testing.T) {
	ctx := context.Background()
	stg := storage.NewMemStorage()

	key := entities.MetricsKey{
		Type: entities.MetricGauge,
		Name: "cpu_usage",
	}

	// Test Set
	value := 0.5
	metric := entities.Metric{
		MetricsKey: key,
		Value:      &value,
	}
	err := stg.Set(ctx, metric)
	require.NoError(t, err)

	// Test Get
	retrievedMetric, ok, err := stg.Get(ctx, metric.MetricsKey)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, metric, retrievedMetric)

	// Test All
	allMetrics, err := stg.All(ctx)
	require.NoError(t, err)
	require.Len(t, allMetrics, 1)
	require.Equal(t, metric, allMetrics[0])
}
