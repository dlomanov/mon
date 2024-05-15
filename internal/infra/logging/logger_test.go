package logging_test

import (
	"fmt"
	"github.com/dlomanov/mon/internal/infra/logging"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithLevel(t *testing.T) {
	tests := []struct {
		name  string
		level string
		want  string
	}{
		{
			name:  "success case",
			level: "debug",
			want:  "debug",
		},
		{
			name:  "success case",
			level: "info",
			want:  "info",
		},
		{
			name:  "success case",
			level: "warn",
			want:  "warn",
		},
		{
			name:  "success case",
			level: "error",
			want:  "error",
		},
		{
			name:  "success case",
			level: "fatal",
			want:  "fatal",
		},
		{
			name:  "success case",
			level: "",
			want:  "info",
		},
		{
			name:  "error case",
			level: "information",
			want:  "",
		},
		{
			name:  "error case",
			level: "warning",
			want:  "",
		},
	}

	for i, tt := range tests {
		name := fmt.Sprintf("%d_%s", i, tt.name)
		t.Run(name, func(t *testing.T) {
			logger, err := logging.WithLevel(tt.level)
			if tt.want == "" {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			lvl := logger.Level().String()
			require.Equal(t, tt.want, lvl)
		})
	}
}
