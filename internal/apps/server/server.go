package server

import (
	"fmt"
	"github.com/dlomanov/mon/internal/apps/server/handlers"
	"github.com/dlomanov/mon/internal/apps/server/logger"
	"github.com/dlomanov/mon/internal/apps/server/middlewares"
	"github.com/dlomanov/mon/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"io"
	"net/http"
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
	defer func(db io.Closer) { _ = db.Close() }(db)

	r := createRouter(db)
	return http.ListenAndServe(cfg.Addr, r)
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
