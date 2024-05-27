// Package server provides functionality to start and manage a web server.
// It includes features for setting up routes, handling HTTP requests,
// and gracefully shutting down the server.
package server

import (
	"context"
	grpcv1 "github.com/dlomanov/mon/internal/apps/server/entrypoints/grpc/v1"
	httpv1 "github.com/dlomanov/mon/internal/apps/server/entrypoints/http/v1"
	"github.com/dlomanov/mon/internal/infra/grpcserver"
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

	httpserv := startHTTPServer(c)
	grpcserv := startGRPCServer(c)
	wait(ctx, c, httpserv, grpcserv)
	shutdownHTTPServer(c, httpserv)
	shutdownGRPCServer(c, grpcserv)

	return nil
}

func startHTTPServer(c *container.Container) *httpserver.Server {
	r := chi.NewRouter()
	httpv1.UseEndpoints(r, c)
	s := httpserver.New(r,
		httpserver.Addr(c.Config.Addr),
		httpserver.ShutdownTimeout(15*time.Second))
	c.Logger.Debug("HTTP-server started")
	return s
}

func startGRPCServer(c *container.Container) *grpcserver.Server {
	s := grpcserver.New(
		grpcserver.Addr(c.Config.GRPCAddr),
		grpcserver.ShutdownTimeout(15*time.Second),
		grpcv1.GetServerOptions(c),
	)
	grpcv1.UseServices(s, c)
	c.Logger.Debug("gRPC-server started")
	return s
}

func wait(
	ctx context.Context,
	c *container.Container,
	httpserv *httpserver.Server,
	grpcserv *grpcserver.Server,
) {
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		c.Logger.Info("cached cancellation", zap.Error(ctx.Err()))
	case s := <-terminate:
		c.Logger.Info("cached terminate signal", zap.String("signal", s.String()))
	case err := <-httpserv.Notify():
		c.Logger.Error("HTTP-server notified error", zap.Error(err))
	case err := <-grpcserv.Notify():
		c.Logger.Error("gRPC-server notified error", zap.Error(err))
	}
}

func shutdownHTTPServer(c *container.Container, s *httpserver.Server) {
	c.Logger.Debug("HTTP-server shutdown")
	if err := s.Shutdown(); err != nil {
		c.Logger.Error("HTTP-server shutdown error", zap.Error(err))
		return
	}
	c.Logger.Debug("HTTP-server shutdown - ok")
}

func shutdownGRPCServer(c *container.Container, s *grpcserver.Server) {
	c.Logger.Debug("gRPC-server shutdown")
	if err := s.Shutdown(); err != nil {
		c.Logger.Error("gRPC-server shutdown error", zap.Error(err))
		return
	}
	c.Logger.Debug("gRPC-server shutdown - ok")
}
