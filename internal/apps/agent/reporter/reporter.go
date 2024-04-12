package reporter

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dlomanov/mon/internal/apps/shared/apimodels"
	"github.com/dlomanov/mon/internal/apps/shared/hashing"
	"github.com/dlomanov/mon/internal/entities"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type Reporter struct {
	logger            *zap.Logger
	client            *resty.Client
	hashKey           string
	workerQueue       chan map[string]entities.Metric
	workerQueueClosed atomic.Bool
	workerCount       uint64
}

func NewReporter(cfg Config, logger *zap.Logger) *Reporter {
	return &Reporter{
		workerCount: cfg.RateLimit,
		workerQueue: make(chan map[string]entities.Metric, cfg.RateLimit),
		client: createClient(cfg.Addr).
			SetRetryWaitTime(1 * time.Second).
			SetRetryMaxWaitTime(5 * time.Second).
			SetRetryCount(3),
		logger:  logger,
		hashKey: cfg.Key,
	}
}

func createClient(addr string) *resty.Client {
	if !strings.HasPrefix(addr, "http") { // ensure protocol schema
		addr = "http://" + addr
	}
	client := resty.New()
	client.SetBaseURL(addr)
	return client
}

func (r *Reporter) Close() {
	if r.workerQueueClosed.CompareAndSwap(false, true) {
		close(r.workerQueue)
	}
}

func (r *Reporter) Enqueue(metrics map[string]entities.Metric) {
	if r.workerQueueClosed.Load() {
		return
	}
	r.workerQueue <- metrics
}

func (r *Reporter) StartWorkers(ctx context.Context) {
	worker := func(number uint64) {
		defer r.logger.Debug("worker stopped", zap.Uint64("worker_number", number), zap.Error(ctx.Err()))

		for ctx.Err() == nil {
			select {
			case <-ctx.Done():
				continue
			case v, open := <-r.workerQueue:
				if !open {
					r.logger.Debug("input queue closed, stopping worker", zap.Uint64("worker_number", number))
					return
				}
				r.report(ctx, v)
				r.logger.Debug("metric reported", zap.Uint64("worker_number", number))
			}
		}
	}

	for i := uint64(0); i < r.workerCount; i++ {
		go worker(i + 1)
	}

	r.logger.Debug("worker started", zap.Uint64("worker_count", r.workerCount))
}

func (r *Reporter) report(ctx context.Context, metrics map[string]entities.Metric) {
	if len(metrics) == 0 {
		return
	}

	data := make([]apimodels.Metric, 0, len(metrics))
	for _, v := range metrics {
		model := apimodels.MapToModel(v)
		data = append(data, model)
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		r.logger.Error("marshaling failed", zap.Error(err))
		return
	}

	compressedJSON, err := compress(dataJSON)
	if err != nil {
		r.logger.Error("compression failed", zap.Error(err))
		return
	}

	request := r.client.
		R().
		SetContext(ctx)
	if r.hashKey != "" {
		hash := hashing.ComputeBase64URLHash(r.hashKey, dataJSON)
		request = request.SetHeader(hashing.HeaderHash, hash)
	}
	_, err = request.
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(compressedJSON).
		Post("/updates/")

	if err != nil {
		r.logger.Error("reporting metrics failed", zap.Error(err))
		return
	}

	r.logger.Debug("metrics reported")
}

func compress(dataJSON []byte) ([]byte, error) {
	buf := bytes.Buffer{}
	cw := gzip.NewWriter(&buf)

	_, err := cw.Write(dataJSON)
	if err != nil {
		return nil, err
	}
	err = cw.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
