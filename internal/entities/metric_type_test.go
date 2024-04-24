package entities_test

import (
	"testing"

	"github.com/dlomanov/mon/internal/entities"
	"github.com/stretchr/testify/require"
)

func TestNewMetricType(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "1 success case",
			key:     string(entities.MetricCounter) + "_test-key",
			wantErr: false,
		},
		{
			name:    "2 success case",
			key:     string(entities.MetricGauge) + "_test-key",
			wantErr: false,
		},
		{
			name:    "5 fail case",
			key:     string(entities.MetricCounter) + "test-key",
			wantErr: true,
		},

		{
			name:    "5 fail case",
			key:     "unknown_test-key",
			wantErr: true,
		},
		{
			name:    "5 fail case",
			key:     "",
			wantErr: true,
		},
		{
			name:    "5 fail case",
			key:     "_",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := entities.NewMetricsKey(tt.key)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.key, key.String())
		})
	}
}
