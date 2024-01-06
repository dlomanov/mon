package logger

import "go.uber.org/zap"

var Log = zap.NewNop()

func WithLevel(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	logger, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = logger
	return nil
}
