package main

import (
	"github.com/dlomanov/mon/internal/handlers/metrics"
	"github.com/dlomanov/mon/internal/handlers/metrics/counter"
	"github.com/dlomanov/mon/internal/handlers/metrics/gauge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestMon_UpdateGauge(t *testing.T) {
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
			storage := make(map[string]metrics.Metric, 1)
			m := &Mon{metrics: storage, logger: log.Default()}

			for _, value := range tt.args.values {
				m.UpdateGauge(tt.args.name, value)
			}

			res, ok := storage[gauge.Metric{Name: tt.args.name}.Key()]
			require.True(t, ok)
			require.IsType(t, gauge.Metric{}, res)
			assert.Equal(t, tt.want, res.(gauge.Metric).Value)
		})
	}
}

func TestMon_UpdateCounter(t *testing.T) {
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
			storage := make(map[string]metrics.Metric, 1)
			m := &Mon{metrics: storage, logger: log.Default()}

			for _, value := range tt.args.values {
				m.UpdateCounter(tt.args.name, value)
			}

			res, ok := storage[counter.Metric{Name: tt.args.name}.Key()]
			require.True(t, ok)
			require.IsType(t, counter.Metric{}, res)
			assert.Equal(t, tt.want, res.(counter.Metric).Value)
		})
	}
}
