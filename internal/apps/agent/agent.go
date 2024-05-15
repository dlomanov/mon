package agent

import (
	"context"
	"github.com/dlomanov/mon/internal/infra/logging"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dlomanov/mon/internal/apps/agent/jobs"
	"github.com/dlomanov/mon/internal/apps/agent/reporter"
	"go.uber.org/zap"
)

const terminateTimeout = time.Second * 3

func Run(cfg Config) (err error) {
	logger, err := logging.WithLevel(cfg.LogLevel)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r, err := reporter.NewReporter(cfg.ReporterConfig, logger, nil)
	if err != nil {
		logger.Error("failed to create reporter", zap.Error(err))
		return err
	}
	defer r.Close()
	r.StartWorkers(ctx)

	go jobs.CollectMetrics(ctx, cfg.CollectorConfig, logger, r.Enqueue)
	go jobs.CollectAdvancedMetrics(ctx, cfg.CollectorConfig, logger, r.Enqueue)

	<-catchTerminate(logger, func() { cancel() })
	logger.Debug("agent stopped")
	return nil
}

func catchTerminate(logger *zap.Logger, onTerminate func()) chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)

		terminate := make(chan os.Signal, 1)

		signal.Notify(terminate,
			syscall.SIGINT,
			syscall.SIGQUIT,
			syscall.SIGTERM)

		s := <-terminate
		logger.Debug("Got one of stop signals, shutting down server gracefully", zap.String("SIGNAL NAME", s.String()))
		onTerminate()
		time.Sleep(terminateTimeout)
	}()
	return done
}
