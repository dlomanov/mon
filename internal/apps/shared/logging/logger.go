package logging

import "go.uber.org/zap"

func WithLevel(level string) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
