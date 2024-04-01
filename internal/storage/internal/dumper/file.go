package dumper

import (
	"encoding/json"
	"io"
	"os"
	"sync"

	"github.com/dlomanov/mon/internal/entities"
	"github.com/dlomanov/mon/internal/storage/internal/mem"
	"go.uber.org/zap"
)

func NewFileDumper(logger *zap.Logger, filePath string) *FileDumper {
	return &FileDumper{
		logger:   logger,
		filePath: filePath,
		mu:       sync.Mutex{},
	}
}

type FileDumper struct {
	logger   *zap.Logger
	filePath string
	mu       sync.Mutex
}

func (f *FileDumper) Load(dest *mem.Storage) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := os.OpenFile(f.filePath, os.O_RDONLY, 0o666)
	if os.IsNotExist(err) {
		f.logger.Debug("file doesn't exist", zap.Error(err))
		return nil
	}
	if err != nil {
		f.logger.Error("failed to open file", zap.Error(err))
		return err
	}
	defer func(file *os.File) { _ = file.Close() }(file)

	m := make(mem.Storage)
	dec := json.NewDecoder(file)

	for {
		data := metric{}
		if err = dec.Decode(&data); err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		entity := entities.Metric{
			MetricsKey: entities.MetricsKey{
				Name: data.Name,
				Type: entities.MustParseMetricType(data.Type),
			},
			Value: data.Value,
			Delta: data.Delta,
		}

		m[entity.String()] = entity
	}

	*dest = m
	f.logger.Debug("metrics loaded")
	return nil
}

func (f *FileDumper) Dump(source mem.Storage) error {
	if len(source) == 0 {
		f.logger.Debug("nothing to dump")
		return nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := os.OpenFile(f.filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666)
	if err != nil {
		f.logger.Error("failed to open file", zap.Error(err))
		return err
	}
	defer func(file *os.File) { _ = file.Close() }(file)

	enc := json.NewEncoder(file)
	for k, v := range source {
		data := metric{
			Name:  v.Name,
			Type:  string(v.Type),
			Delta: v.Delta,
			Value: v.Value,
		}
		valueStr := v.StringValue()

		err = enc.Encode(data)
		if err != nil {
			f.logger.Error("failed to encode metric",
				zap.String("key", k),
				zap.String("value", valueStr))
			return err
		}
	}

	f.logger.Debug("metrics dumped")
	return nil
}

type metric struct {
	Name  string   `json:"name"`
	Type  string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}
