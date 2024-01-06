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
	"net/http"
)

func Run(cfg Config) error {
	db := storage.NewStorage()
	r := createRouter(db)
	err := logger.WithLevel(cfg.LogLevel)
	if err != nil {
		return err
	}

	cfgStr := fmt.Sprintf("%+v", cfg)
	logger.Log.Info("server running...", zap.String("cfg", cfgStr))

	return http.ListenAndServe(cfg.Addr, r)
}

func createRouter(db storage.Storage) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middlewares.Logger)
	router.Use(middleware.Recoverer)
	router.Post("/update/{type}/{name}/{value}", handlers.UpdateByParams(db))
	router.Post("/update/", handlers.UpdateByJSON(db))
	router.Get("/value/{type}/{name}", handlers.GetByParams(db))
	router.Post("/value/", handlers.GetByJSON(db))
	router.Get("/", handlers.Report(db))

	return router
}
