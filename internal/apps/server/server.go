// Package server provides functionality to start and manage a web server.
// It includes features for setting up routes, handling HTTP requests,
// and gracefully shutting down the server.
package server

import (
	"context"
	v1 "github.com/dlomanov/mon/internal/apps/server/entrypoints/http/v1"
	"github.com/dlomanov/mon/internal/infra/httpserver"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dlomanov/mon/internal/apps/server/container"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Run - starts the server with the provided configuration.
// It wires the dependencies, sets up the router, and starts the server.
// It also handles graceful shutdown.
func Run(ctx context.Context, cfg Config, logger *zap.Logger) error {
	c, err := container.NewContainer(ctx, logger, cfg.Config)
	if err != nil {
		return err
	}
	defer c.Close()

	s := startServer(c)
	wait(ctx, c, s)
	shutdownServer(c, s)

	return nil
}

func startServer(c *container.Container) *httpserver.Server {
	r := chi.NewRouter()
	v1.UseEndpoints(r, c)
	s := httpserver.New(r,
		httpserver.Addr(c.Config.Addr),
		httpserver.ShutdownTimeout(15*time.Second))
	c.Logger.Debug("server started")

	return s
}

func wait(
	ctx context.Context,
	c *container.Container,
	server *httpserver.Server,
) {
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		c.Logger.Info("cached cancellation", zap.Error(ctx.Err()))
	case s := <-terminate:
		c.Logger.Info("cached terminate signal", zap.String("signal", s.String()))
	case err := <-server.Notify():
		c.Logger.Error("server notified error", zap.Error(err))
	}
}

func shutdownServer(c *container.Container, s *httpserver.Server) {
	c.Logger.Debug("server shutdown")
	if err := s.Shutdown(); err != nil {
		c.Logger.Error("server shutdown error", zap.Error(err))
		return
	}
	c.Logger.Debug("server shutdown - ok")
}
