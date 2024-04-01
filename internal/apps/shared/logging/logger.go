package logging

import "go.uber.org/zap"

// WithLevel creates a new zap.Logger with the specified logging level.
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
