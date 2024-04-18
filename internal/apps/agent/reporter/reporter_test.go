package reporter_test

import (
	"context"
	"testing"
	"time"

	"github.com/dlomanov/mon/internal/apps/agent/reporter"
	"github.com/dlomanov/mon/internal/entities"
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestReporter(t *testing.T) {
	const addr = "http://localhost:8089"
	const url = addr + "/updates/"

	cfg := reporter.Config{
		Addr:      addr,
		Key:       "test_key",
		RateLimit: 1,
	}
	client := resty.New()
	r := reporter.NewReporter(cfg, zaptest.NewLogger(t), client)
	defer r.Close()

	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", url, httpmock.NewStringResponder(200, `{}`))

	value := float64(0.5)
	assert.NotPanics(t, func() {
		r.StartWorkers(context.Background())
		r.Enqueue(map[string]entities.Metric{
			"gauge_test": {
				MetricsKey: entities.MetricsKey{
					Type: entities.MetricGauge,
					Name: "test",
				},
				Value: &value,
			},
		})
	})
	r.Close()
	time.Sleep(1 * time.Second)

	callCount := httpmock.GetTotalCallCount()
	assert.Equal(t, 1, callCount)
	info := httpmock.GetCallCountInfo()
	assert.Equal(t, 1, info["POST "+url])
}
