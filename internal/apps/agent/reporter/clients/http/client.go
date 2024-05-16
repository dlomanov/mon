package http

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/dlomanov/mon/internal/apps/agent/reporter"
	"github.com/dlomanov/mon/internal/apps/agent/reporter/utils"
	"github.com/dlomanov/mon/internal/apps/shared/apimodels"
	"github.com/dlomanov/mon/internal/entities"
	"github.com/dlomanov/mon/internal/infra/services/encrypt"
	"github.com/dlomanov/mon/internal/infra/services/hashing"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"os"
	"strings"
	"time"
)

var _ reporter.Client = (*Client)(nil)

type (
	Client struct {
		client  *resty.Client
		logger  *zap.Logger
		enc     *encrypt.Encryptor
		hashKey string
	}
	Config struct {
		Addr          string
		PublicKeyPath string
		HashKey       string
	}
)

func (r *Client) Close() error {
	return nil
}

func New(
	logger *zap.Logger,
	config Config,
	client *resty.Client,
) (*Client, error) {
	enc, err := createEncryptor(config.PublicKeyPath)
	if err != nil {
		return nil, err
	}
	return &Client{
		logger: logger,
		client: createClient(client, config.Addr).
			SetRetryWaitTime(1 * time.Second).
			SetRetryMaxWaitTime(5 * time.Second).
			SetRetryCount(3),
		enc:     enc,
		hashKey: config.HashKey,
	}, nil
}

func (r *Client) Report(ctx context.Context, metrics map[string]entities.Metric) {
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

	ip, err := utils.GetOutboundIP()
	if err != nil {
		r.logger.Error("get outbound ip failed", zap.Error(err))
	} else {
		headers["X-Real-IP"] = ip.String()
	}

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

func (r *Client) encrypt(input []byte) ([]byte, bool, error) {
	if r.enc == nil {
		return input, false, nil
	}
	output, err := r.enc.Encrypt(input)
	return output, true, err
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
