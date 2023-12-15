package server

import (
	"github.com/dlomanov/mon/internal/apps/server/handlers"
	"github.com/dlomanov/mon/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func ListenAndServe(addr string) error {
	db := storage.NewStorage()
	r := createRouter(db)
	return http.ListenAndServe(addr, r)
}

func createRouter(db storage.Storage) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Post("/update/{type}/{name}/{value}", handlers.Update(db))
	router.Get("/value/{type}/{name}", handlers.Get(db))
	router.Get("/", handlers.Report(db))

	return router
}
