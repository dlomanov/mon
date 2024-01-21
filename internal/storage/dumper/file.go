package dumper

import (
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"os"
	"sync"
)

func init() {
	var _ Dumper = (*FileDumper)(nil)
}

func NewFileDumper(
	logger *zap.Logger,
	filePath string,
	restore bool,
) *FileDumper {
	return &FileDumper{
		logger:   logger,
		filePath: filePath,
		restore:  restore,
		mu:       sync.Mutex{},
	}
}

type FileDumper struct {
	logger   *zap.Logger
	filePath string
	restore  bool
	mu       sync.Mutex
}

func (f *FileDumper) Load(dest *map[string]string) error {
	if !f.restore {
		f.logger.Debug("load disabled")
		return nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := os.OpenFile(f.filePath, os.O_RDONLY, 0666)
	if os.IsNotExist(err) {
		f.logger.Debug("file doesn't exist", zap.Error(err))
		return nil
	}
	if err != nil {
		f.logger.Error("failed to open file", zap.Error(err))
		return err
	}
	defer func(file *os.File) { _ = file.Close() }(file)

	m := make(map[string]string)
	data := pair{}
	dec := json.NewDecoder(file)

	for {
		if err = dec.Decode(&data); err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		m[data.Key] = data.Value
		f.logger.Debug("- load metric",
			zap.String("key", data.Key),
			zap.String("value", data.Value))
	}

	*dest = m
	f.logger.Debug("metrics loaded")
	return nil
}

func (f *FileDumper) Dump(source map[string]string) error {
	if len(source) == 0 {
		f.logger.Debug("nothing to dump")
		return nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := os.OpenFile(f.filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		f.logger.Error("failed to open file", zap.Error(err))
		return err
	}
	defer func(file *os.File) { _ = file.Close() }(file)

	enc := json.NewEncoder(file)
	for k, v := range source {
		data := pair{Key: k, Value: v}
		err = enc.Encode(data)
		if err != nil {
			f.logger.Error("failed to encode metric",
				zap.String("key", k),
				zap.String("value", v))
			return err
		}
		f.logger.Debug("- dump metric",
			zap.String("key", data.Key),
			zap.String("value", data.Value))
	}

	f.logger.Debug("metrics dumped")
	return nil
}

type pair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
