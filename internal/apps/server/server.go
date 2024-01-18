package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/dlomanov/mon/internal/apps/server/handlers"
	"github.com/dlomanov/mon/internal/apps/server/middlewares"
	"github.com/dlomanov/mon/internal/storage"
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

	container, err := configureServices(ctx, cfg)
	if err != nil {
		return err
	}
	stg := container.MemStorage
	logger := container.Logger

	logger.Info("server running...", zap.String("cfg", cfgStr))

	server := &http.Server{Addr: cfg.Addr, Handler: createRouter(container)}
	go dumpLoop(ctx, stg, logger)
	go catchTerminate(server, logger, func() {
		cancel()
		_ = stg.Close()
	})
	return server.ListenAndServe()
}

func createRouter(container *serviceContainer) *chi.Mux {
	stg := container.Storage
	logger := container.Logger

	router := chi.NewRouter()
	router.Use(middlewares.Logger(logger))
	router.Use(middlewares.Compressor)
	router.Use(middleware.Recoverer)
	router.Post("/update/{type}/{name}/{value}", handlers.UpdateByParams(logger, stg))
	router.Post("/update/", handlers.UpdateByJSON(logger, stg))
	router.Get("/value/{type}/{name}", handlers.GetByParams(logger, stg))
	router.Post("/value/", handlers.GetByJSON(logger, stg))
	router.Get("/", handlers.Report(logger, stg))

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

func dumpLoop(
	ctx context.Context,
	stg *storage.MemStorage,
	logger *zap.Logger,
) {
	err := stg.DumpLoop(ctx)
	if errors.Is(err, context.Canceled) {
		logger.Debug("dump loop cancelled", zap.Error(err))
		return
	}
	if err != nil {
		logger.Error("failed dump loop", zap.Error(err))
	}
}
