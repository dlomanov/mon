package server

import (
	"context"
	"fmt"
	"github.com/dlomanov/mon/internal/apps/server/handlers"
	"github.com/dlomanov/mon/internal/apps/server/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Run(cfg Config) error {
	cfgStr := fmt.Sprintf("%+v", cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	container, err := handlers.NewContainer(ctx, cfg.Config)
	if err != nil {
		return err
	}
	defer func(container *handlers.Container) { _ = container.Close() }(container)

	logger := container.Logger
	logger.Info("server running...", zap.String("cfg", cfgStr))
	server := &http.Server{Addr: cfg.Addr, Handler: createRouter(container)}
	go catchTerminate(server, logger, func() { cancel() })
	return server.ListenAndServe()
}

func createRouter(container *handlers.Container) *chi.Mux {
	logger := container.Logger

	router := chi.NewRouter()
	router.Use(middlewares.Logger(logger))
	router.Use(middlewares.Compressor)
	router.Use(middleware.Recoverer)
	router.Post("/update/{type}/{name}/{value}", container.UpdateByParams())
	router.Post("/update/", container.UpdateByJSON())
	router.Post("/updates/", container.UpdatesByJSON())
	router.Get("/value/{type}/{name}", container.GetByParams())
	router.Post("/value/", container.GetByJSON())
	router.Get("/ping", container.PingDB())
	router.Get("/", container.Report())

	return router
}

func catchTerminate(
	server *http.Server,
	logger *zap.Logger,
	onTerminate func(),
) {
	terminate := make(chan os.Signal, 1)

	signal.Notify(terminate,
		syscall.SIGINT,
		syscall.SIGTERM)

	s := <-terminate
	logger.Debug("Got one of stop signals, shutting down server gracefully", zap.String("SIGNAL NAME", s.String()))
	onTerminate()

	err := server.Shutdown(context.Background())
	logger.Error("Error from shutdown", zap.Error(err))
}
