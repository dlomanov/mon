package server

import (
	"context"
	"fmt"
	"github.com/dlomanov/mon/internal/apps/server/handlers"
	"github.com/dlomanov/mon/internal/apps/server/logger"
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
	err := logger.WithLevel(cfg.LogLevel)
	if err != nil {
		return err
	}

	cfgStr := fmt.Sprintf("%+v", cfg)
	logger.Log.Info("server running...", zap.String("cfg", cfgStr))

	db := storage.NewMemStorage(
		logger.Log,
		storage.Config{
			StoreInterval:   cfg.StoreInterval,
			FileStoragePath: cfg.FileStoragePath,
			Restore:         cfg.Restore,
		})

	server := &http.Server{Addr: cfg.Addr, Handler: createRouter(db)}

	go dumpLoop(db, logger.Log)

	go catchTerminate(server, db, logger.Log)

	return server.ListenAndServe()
}

func createRouter(db storage.Storage) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middlewares.Logger)
	router.Use(middlewares.Compressor)
	router.Use(middleware.Recoverer)
	router.Post("/update/{type}/{name}/{value}", handlers.UpdateByParams(db))
	router.Post("/update/", handlers.UpdateByJSON(db))
	router.Get("/value/{type}/{name}", handlers.GetByParams(db))
	router.Post("/value/", handlers.GetByJSON(db))
	router.Get("/", handlers.Report(db))

	return router
}

func catchTerminate(
	server *http.Server,
	db *storage.MemStorage,
	logger *zap.Logger,
) {
	terminate := make(chan os.Signal, 1)

	signal.Notify(terminate,
		syscall.SIGINT,
		syscall.SIGTERM)

	s := <-terminate
	_ = db.Close()
	logger.Info("Got one of stop signals, shutting down server gracefully", zap.String("SIGNAL NAME", s.String()))

	err := server.Shutdown(context.Background())
	logger.Info("Error from shutdown", zap.Error(err))
}

func dumpLoop(db *storage.MemStorage, logger *zap.Logger) {
	err := db.DumpLoop()
	if err != nil {
		logger.Error("failed dump loop", zap.Error(err))
	}
}
