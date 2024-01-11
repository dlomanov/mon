package storage

import (
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"os"
)

func load(mem *MemStorage) error {
	path := mem.config.FileStoragePath

	if path == "" {
		mem.logger.Debug("no file path specified")
		return nil
	}

	if !mem.config.Restore {
		mem.logger.Debug("load disabled")
		return nil
	}

	mem.mu.Lock()
	defer mem.mu.Unlock()

	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if os.IsNotExist(err) {
		mem.logger.Debug("file doesn't exist", zap.Error(err))
		return nil
	}
	if err != nil {
		mem.logger.Error("failed to open file", zap.Error(err))
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
		mem.logger.Debug("- load metric",
			zap.String("key", data.Key),
			zap.String("value", data.Value))
	}

	mem.storage = m
	mem.logger.Debug("metrics loaded")
	return nil
}

func dump(mem *MemStorage) error {
	if !canDump(mem) {
		return nil
	}

	file, err := os.OpenFile(mem.config.FileStoragePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		mem.logger.Error("failed to open file", zap.Error(err))
		return err
	}
	defer func(file *os.File) { _ = file.Close() }(file)

	enc := json.NewEncoder(file)
	for k, v := range mem.storage {
		data := pair{Key: k, Value: v}
		err = enc.Encode(data)
		if err != nil {
			mem.logger.Error("failed to encode metric",
				zap.String("key", k),
				zap.String("value", v))
			return err
		}
		mem.logger.Debug("- dump metric",
			zap.String("key", data.Key),
			zap.String("value", data.Value))
	}

	mem.logger.Debug("metrics dumped")
	return nil
}

func canDump(mem *MemStorage) bool {
	path := mem.config.FileStoragePath

	if path == "" {
		mem.logger.Debug("dump disabled")
		return false
	}

	if len(mem.storage) == 0 {
		mem.logger.Debug("nothing to dump")
		return false
	}

	return true
}

type pair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
