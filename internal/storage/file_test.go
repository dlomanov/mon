package storage_test

import (
	"context"
	"os"
	"testing"

	"github.com/dlomanov/mon/internal/entities"
	"github.com/dlomanov/mon/internal/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestFileStorage(t *testing.T) {
	logger := zaptest.NewLogger(t)

	dir := t.TempDir()
	file, err := os.CreateTemp(dir, "test")
	require.NoError(t, err)
	temppath := file.Name()
	file.Close()
	require.NoError(t, err)

	fs, err := storage.NewFileStorage(logger,
		storage.FileStorageConfig{
			StoreInterval:   0,
			FileStoragePath: temppath,
			Restore:         false,
		})
	require.NoError(t, err)
	defer func(fs *storage.FileStorage) {
		require.NoError(t, fs.Close())
	}(fs)

	// Test Set
	ctx := context.Background()
	key := entities.MetricsKey{
		Type: entities.MetricGauge,
		Name: "cpu_usage",
	}
	value := 0.5
	metric := entities.Metric{
		MetricsKey: key,
		Value:      &value,
	}
	err = fs.Set(ctx, metric)
	require.NoError(t, err)

	// Test Get
	retrievedMetric, ok, err := fs.Get(ctx, metric.MetricsKey)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, metric, retrievedMetric)

	// Test All after restore
	require.NoError(t, fs.Close())
	fs, err = storage.NewFileStorage(logger,
		storage.FileStorageConfig{
			StoreInterval:   0,
			FileStoragePath: temppath,
			Restore:         true,
		})
	require.NoError(t, err)
	defer func(fs *storage.FileStorage) {
		require.NoError(t, fs.Close())
	}(fs)
	allMetrics, err := fs.All(ctx)
	require.NoError(t, err)
	require.Len(t, allMetrics, 1)
	require.Equal(t, metric, allMetrics[0])
}
