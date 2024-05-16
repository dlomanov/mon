package agent

import (
	"context"
	"github.com/dlomanov/mon/internal/apps/agent/reporter"
	grpcclient "github.com/dlomanov/mon/internal/apps/agent/reporter/clients/grpc"
	httpclient "github.com/dlomanov/mon/internal/apps/agent/reporter/clients/http"
	"github.com/dlomanov/mon/internal/infra/logging"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dlomanov/mon/internal/apps/agent/jobs"
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

	rc, err := createReportClient(logger, cfg)
	if err != nil {
		logger.Error("failed to create report client", zap.Error(err))
		return err
	}
	r := reporter.NewReporter(logger, cfg.RateLimit, rc)
	defer r.Close()

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

func createReportClient(logger *zap.Logger, cfg Config) (reporter.Client, error) {
	if cfg.GRPCAddr != "" {
		return grpcclient.New(logger, cfg.GRPCAddr)
	}
	return httpclient.New(logger, httpclient.Config{
		Addr:          cfg.Addr,
		PublicKeyPath: cfg.PublicKeyPath,
		HashKey:       cfg.HashKey,
	}, nil)
}
