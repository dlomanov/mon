package entities_test

import (
	"testing"

	"github.com/dlomanov/mon/internal/entities"
	"github.com/stretchr/testify/require"
)

func TestNewMetric(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
	}{
		{
			name:    "1 success case",
			key:     string(entities.MetricCounter) + "_test-key",
			value:   "1",
			wantErr: false,
		},
		{
			name:    "2 success case",
			key:     string(entities.MetricGauge) + "_test-key",
			value:   "1.1",
			wantErr: false,
		},
		{
			name:    "3 success case",
			key:     string(entities.MetricGauge) + "_test-key",
			value:   "1",
			wantErr: false,
		},
		{
			name:    "4 fail case",
			key:     string(entities.MetricCounter) + "_test-key",
			value:   "1.1",
			wantErr: true,
		},
		{
			name:    "5 fail case",
			key:     string(entities.MetricCounter) + "_test-key",
			value:   "1.1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := entities.NewMetricsKey(tt.key)
			require.NoError(t, err)

			m, err := entities.NewMetric(key, tt.value)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.Equal(t, tt.key, key.String())
			require.Equal(t, tt.value, m.StringValue())
		})
	}
}
