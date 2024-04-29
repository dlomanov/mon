// Package server provides functionality to start and manage a web server.
// It includes features for setting up routes, handling HTTP requests,
// and gracefully shutting down the server.
package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	_ "github.com/dlomanov/mon/internal/apps/server/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/dlomanov/mon/internal/apps/server/container"
	"github.com/dlomanov/mon/internal/apps/server/handlers"
	"github.com/dlomanov/mon/internal/apps/server/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// Run - starts the server with the provided configuration.
// It wiring the dependencies, sets up the router, and starts the server.
// It also handles graceful shutdown.
//
//	@title		mon API
//	@version	1.0
func Run(ctx context.Context, cfg Config, logger *zap.Logger) error {
	cfgStr := fmt.Sprintf("%+v", cfg)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	c, err := container.NewContainer(ctx, logger, cfg.Config)
	if err != nil {
		return err
	}
	defer func(c *container.Container) { _ = c.Close() }(c)

	logger.Info("server running...", zap.String("cfg", cfgStr))
	server := &http.Server{Addr: cfg.Addr, Handler: createRouter(c)}
	go catchTerminate(server, logger, func() { cancel() })
	return server.ListenAndServe()
}

func createRouter(container *container.Container) *chi.Mux {
	logger := container.Logger

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(middlewares.Logger(logger))
	router.Use(middlewares.Compressor)
	router.Use(middlewares.Decrypter(logger, container.Dec))
	router.Use(middlewares.Hash(container))
	useSwagger(router, container.Config)
	router.Post("/update/{type}/{name}/{value}", handlers.UpdateByParams(container))
	router.Post("/update/", handlers.UpdateByJSON(container))
	router.Post("/updates/", handlers.UpdatesByJSON(container))
	router.Get("/value/{type}/{name}", handlers.GetByParams(container))
	router.Post("/value/", handlers.GetByJSON(container))
	router.Get("/ping", handlers.PingDB(container))
	router.Get("/", handlers.Report(container))

	return router
}

func useSwagger(router *chi.Mux, cfg container.Config) {
	url := cfg.Addr
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}
	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(url+"/swagger/doc.json"),
	))
}

func catchTerminate(
	server *http.Server,
	logger *zap.Logger,
	onTerminate func(),
) {
	terminate := make(chan os.Signal, 1)

	signal.Notify(terminate,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)

	s := <-terminate
	logger.Debug("Got one of stop signals, shutting down server gracefully", zap.String("SIGNAL NAME", s.String()))
	onTerminate()

	err := server.Shutdown(context.Background())
	logger.Error("Error from shutdown", zap.Error(err))
}
