package reporter_test

import (
	"github.com/dlomanov/mon/internal/apps/agent/reporter"
	httpclient "github.com/dlomanov/mon/internal/apps/agent/reporter/clients/http"
	"github.com/dlomanov/mon/internal/apps/agent/reporter/utils"
	"github.com/dlomanov/mon/internal/entities"
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"testing"
	"time"
)

func TestReporter(t *testing.T) {
	const addr = "http://localhost:8089"
	const url = addr + "/updates/"

	client := resty.New()
	rc, err := httpclient.New(
		zaptest.NewLogger(t),
		httpclient.Config{
			Addr:    addr,
			HashKey: "test_key",
		}, client)
	require.NoError(t, err)

	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", url, httpmock.NewStringResponder(200, `{}`))

	r := reporter.NewReporter(zaptest.NewLogger(t), 1, rc)
	require.NoError(t, err)
	defer r.Close()

	value := 0.5
	assert.NotPanics(t, func() {
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
	time.Sleep(1 * time.Second)
	r.Close()

	callCount := httpmock.GetTotalCallCount()
	assert.Equal(t, 1, callCount)
	info := httpmock.GetCallCountInfo()
	assert.Equal(t, 1, info["POST "+url])
}

func TestGetOutboundIP(t *testing.T) {
	localAddr, err := utils.GetOutboundIP()
	require.NoError(t, err)
	t.Log("IP:", localAddr)
}
