package collector

import (
	"github.com/dlomanov/mon/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestCollector_UpdateGauge(t *testing.T) {
	type args struct {
		name   string
		values []float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "success case",
			args: args{
				name:   "test-key",
				values: []float64{1.0001, 2.0002, 3.0003},
			},
			want: 3.0003,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := make(map[string]entities.Metric, 1)
			c := Collector{metrics: storage, logger: log.Default()}

			for _, value := range tt.args.values {
				c.UpdateGauge(tt.args.name, value)
			}

			key := entities.MetricsKey{Name: tt.args.name, Type: entities.MetricGauge}
			res, ok := storage[key.String()]
			require.True(t, ok)
			assert.Equal(t, tt.want, *res.Value)
		})
	}
}

func TestCollector_UpdateCounter(t *testing.T) {
	type args struct {
		name   string
		values []int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "success case",
			args: args{
				name:   "test-key",
				values: []int64{1, 2, 3},
			},
			want: 6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := make(map[string]entities.Metric, 1)
			c := Collector{metrics: storage, logger: log.Default()}

			for _, value := range tt.args.values {
				c.UpdateCounter(tt.args.name, value)
			}

			key := entities.MetricsKey{Name: tt.args.name, Type: entities.MetricCounter}
			res, ok := storage[key.String()]
			require.True(t, ok)
			assert.Equal(t, tt.want, *res.Delta)
		})
	}
}
