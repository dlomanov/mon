package handlers

import (
	"github.com/dlomanov/mon/internal/handlers/apperrors"
	"github.com/dlomanov/mon/internal/handlers/metrics"
	"github.com/dlomanov/mon/internal/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateHandler(t *testing.T) {
	type want struct {
		code int
	}
	type args struct {
		method string
		path   string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success gauge case",
			args: args{method: http.MethodPost, path: "/update/gauge/key/3.0000003"},
			want: want{code: http.StatusOK},
		},
		{
			name: "success counter case",
			args: args{method: http.MethodPost, path: "/update/counter/key/1"},
			want: want{code: http.StatusOK},
		},
		{
			name: "invalid counter value case 1",
			args: args{method: http.MethodPost, path: "/update/counter/key/1.00001"},
			want: want{code: http.StatusBadRequest},
		},
		{
			name: "invalid counter value case 2",
			args: args{method: http.MethodPost, path: "/update/counter/key/"},
			want: want{code: http.StatusBadRequest},
		},
		{
			name: "invalid counter value case 3",
			args: args{method: http.MethodPost, path: "/update/counter/key/none"},
			want: want{code: http.StatusBadRequest},
		},
		{
			name: "invalid gauge value case 1",
			args: args{method: http.MethodPost, path: "/update/counter/key/none"},
			want: want{code: http.StatusBadRequest},
		},
		{
			name: "invalid gauge value case 2",
			args: args{method: http.MethodPost, path: "/update/counter/key/"},
			want: want{code: http.StatusBadRequest},
		},
		{
			name: "invalid path 1",
			args: args{method: http.MethodPost, path: "/"},
			want: want{code: http.StatusNotFound},
		},
		{
			name: "invalid path 2",
			args: args{method: http.MethodPost, path: "/update"},
			want: want{code: http.StatusNotFound},
		},
		{
			name: "invalid path 3",
			args: args{method: http.MethodPost, path: "/update/"},
			want: want{code: http.StatusBadRequest},
		},
		{
			name: "invalid path 4",
			args: args{method: http.MethodPost, path: "/update/counter"},
			want: want{code: http.StatusNotFound},
		},
		{
			name: "invalid path 5",
			args: args{method: http.MethodPost, path: "/update/counter/"},
			want: want{code: http.StatusNotFound},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := storage.NewStorage()
			r := httptest.NewRequest(tt.args.method, tt.args.path, nil)
			w := httptest.NewRecorder()
			h := UpdateHandler(db)

			h.ServeHTTP(w, r)
			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode, "Unexpected status code")
		})
	}
}

func TestHandleGauge(t *testing.T) {
	type args struct {
		values []metrics.Metric
	}
	type want struct {
		successExpected bool
		errCode         apperrors.Code
		expectedValue   string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success case",
			args: args{
				values: []metrics.Metric{
					{
						Type:  metrics.MetricGauge,
						Name:  "metric-key",
						Value: 1.0001,
					},
					{
						Type:  metrics.MetricGauge,
						Name:  "metric-key",
						Value: 2.0001,
					},
				},
			},
			want: want{
				successExpected: true,
				expectedValue:   "2.0001",
			},
		},
		{
			name: "invalid value case",
			args: args{
				values: []metrics.Metric{
					{
						Type:  metrics.MetricGauge,
						Name:  "metric-key",
						Value: int64(1.0),
					},
				},
			},
			want: want{
				successExpected: false,
				errCode:         ErrInvalidMetricValueType,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := storage.NewStorage()

			for _, value := range tt.args.values {
				err := HandleGauge(value, db)

				if tt.want.successExpected {
					assert.Nil(t, err)
					continue
				}

				assert.NotNil(t, err)

				var apperr apperrors.AppError
				assert.ErrorAs(t, err, &apperr)
				assert.Equal(t, tt.want.errCode, apperr.Code)
			}

			if !tt.want.successExpected {
				return
			}

			value, ok := db.Get(tt.args.values[0].Key())
			assert.True(t, ok, "Value not found")
			assert.Equal(t, tt.want.expectedValue, value)
		})
	}
}

func TestHandleCounter(t *testing.T) {
	type args struct {
		values []metrics.Metric
	}
	type want struct {
		successExpected bool
		errCode         apperrors.Code
		expectedValue   string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success case",
			args: args{
				values: []metrics.Metric{
					{
						Type:  metrics.MetricCounter,
						Name:  "metric-key",
						Value: int64(1),
					},
					{
						Type:  metrics.MetricCounter,
						Name:  "metric-key",
						Value: int64(9),
					},
				},
			},
			want: want{
				successExpected: true,
				expectedValue:   "10",
			},
		},
		{
			name: "invalid value case",
			args: args{
				values: []metrics.Metric{
					{
						Type:  metrics.MetricCounter,
						Name:  "metric-key",
						Value: float64(1),
					},
				},
			},
			want: want{
				successExpected: false,
				errCode:         ErrInvalidMetricValueType,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := storage.NewStorage()

			for _, value := range tt.args.values {
				err := HandleCounter(value, db)

				if tt.want.successExpected {
					assert.Nil(t, err)
					continue
				}

				assert.NotNil(t, err)

				var apperr apperrors.AppError
				assert.ErrorAs(t, err, &apperr)
				assert.Equal(t, tt.want.errCode, apperr.Code)
			}

			if !tt.want.successExpected {
				return
			}

			value, ok := db.Get(tt.args.values[0].Key())
			assert.True(t, ok, "Value not found")
			assert.Equal(t, tt.want.expectedValue, value)
		})
	}
}
