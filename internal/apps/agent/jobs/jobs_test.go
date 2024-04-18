package jobs_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/dlomanov/mon/internal/apps/agent/collector"
	"github.com/dlomanov/mon/internal/apps/agent/jobs"
	"github.com/dlomanov/mon/internal/entities"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCollectMetrics(t *testing.T) {
	timeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelTimeout()
	ctx, cancel := context.WithTimeout(timeoutCtx, 1*time.Second)
	defer cancel()

	doneCh := make(chan struct{})

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		jobs.CollectMetrics(
			ctx,
			collector.Config{
				PollInterval:   10 * time.Millisecond,
				ReportInterval: 20 * time.Millisecond,
			},
			zap.NewNop(),
			func(map[string]entities.Metric) {})
		wg.Done()
	}()
	go func() {
		wg.Wait()
		close(doneCh)
	}()

	select {
	case <-doneCh:
	case <-timeoutCtx.Done():
	}

	assert.NoError(t, timeoutCtx.Err())
}
