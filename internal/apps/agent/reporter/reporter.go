package reporter

import (
	"context"
	"github.com/dlomanov/mon/internal/entities"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
)

type (
	ReportQueue struct {
		logger            *zap.Logger
		client            Client
		queue             chan map[string]entities.Metric
		workerQueueClosed atomic.Bool
		workerCount       uint64
		stop              func()
		stopCtx           context.Context
		stopped           chan struct{}
	}
	Client interface {
		Report(ctx context.Context, metrics map[string]entities.Metric)
		Close() error
	}
)

func NewReporter(
	logger *zap.Logger,
	rateLimit uint64,
	client Client,
) *ReportQueue {
	stopCtx, stop := context.WithCancel(context.Background())
	r := &ReportQueue{
		logger:      logger,
		client:      client,
		workerCount: rateLimit,
		queue:       make(chan map[string]entities.Metric, rateLimit),
		stop:        stop,
		stopCtx:     stopCtx,
		stopped:     make(chan struct{}),
	}

	go r.start()

	return r
}

func (r *ReportQueue) Enqueue(metrics map[string]entities.Metric) {
	if r.workerQueueClosed.Load() {
		return
	}
	r.queue <- metrics
}

func (r *ReportQueue) Close() {
	if r.workerQueueClosed.CompareAndSwap(false, true) {
		close(r.queue)
		r.stop()
		<-r.stopped
		if err := r.client.Close(); err != nil {
			r.logger.Error("failed to close client", zap.Error(err))
		}
	}
}

func (r *ReportQueue) start() {
	defer close(r.stopped)

	var wg sync.WaitGroup
	worker := func(number uint64) {
		defer wg.Done()

		for r.stopCtx.Err() == nil {
			select {
			case <-r.stopCtx.Done():
				continue
			case v, open := <-r.queue:
				if !open {
					r.logger.Debug("input queue closed, stopping worker", zap.Uint64("worker_number", number))
					break
				}
				r.client.Report(r.stopCtx, v)
				r.logger.Debug("metric reported", zap.Uint64("worker_number", number))
			}
		}

		r.logger.Debug("worker stopped", zap.Uint64("worker_number", number), zap.Error(r.stopCtx.Err()))
	}

	for i := uint64(0); i < r.workerCount; i++ {
		wg.Add(1)
		go worker(i + 1)
	}
	r.logger.Debug("worker started", zap.Uint64("worker_count", r.workerCount))
	wg.Wait()
	r.logger.Debug("all worker stopped", zap.Uint64("worker_count", r.workerCount))
}
