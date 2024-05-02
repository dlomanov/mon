package reporter

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dlomanov/mon/internal/apps/shared/apimodels"
	"github.com/dlomanov/mon/internal/apps/shared/hashing"
	"github.com/dlomanov/mon/internal/entities"
	"github.com/dlomanov/mon/internal/services/encrypt"
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
	enc               *encrypt.Encryptor
}

func NewReporter(
	cfg Config,
	logger *zap.Logger,
	client *resty.Client,
) (*Reporter, error) {
	enc, err := createEncryptor(cfg.PublicKeyPath)
	if err != nil {
		return nil, err
	}

	return &Reporter{
		workerCount: cfg.RateLimit,
		workerQueue: make(chan map[string]entities.Metric, cfg.RateLimit),
		client: createClient(client, cfg.Addr).
			SetRetryWaitTime(1 * time.Second).
			SetRetryMaxWaitTime(5 * time.Second).
			SetRetryCount(3),
		logger:  logger,
		hashKey: cfg.Key,
		enc:     enc,
	}, nil
}

func createClient(client *resty.Client, addr string) *resty.Client {
	if !strings.HasPrefix(addr, "http") { // ensure protocol schema
		addr = "http://" + addr
	}
	if client == nil {
		client = resty.New()
	}
	client.SetBaseURL(addr)
	return client
}

func createEncryptor(keyPath string) (enc *encrypt.Encryptor, err error) {
	if keyPath == "" {
		return nil, nil
	}
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	enc, err = encrypt.NewEncryptor(key)
	if err != nil {
		return nil, err
	}
	return enc, nil
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

	headers := map[string]string{}
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
	headers["Content-Type"] = "application/json"

	if r.hashKey != "" {
		hash := hashing.ComputeBase64URLHash(r.hashKey, dataJSON)
		headers[hashing.HeaderHash] = hash
	}

	encJSON, encrypted, err := r.encrypt(dataJSON)
	if err != nil {
		r.logger.Error("encription failed", zap.Error(err))
		return
	}
	if encrypted {
		headers["Encryption"] = ""
	}

	compressedJSON, err := compress(encJSON)
	if err != nil {
		r.logger.Error("compression failed", zap.Error(err))
		return
	}
	headers["Content-Encoding"] = "gzip"
	headers["Accept-Encoding"] = "gzip"

	_, err = r.client.
		R().
		SetContext(ctx).
		SetHeaders(headers).
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

func (r *Reporter) encrypt(input []byte) ([]byte, bool, error) {
	if r.enc == nil {
		return input, false, nil
	}
	output, err := r.enc.Encrypt(input)
	return output, true, err
}
