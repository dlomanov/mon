package handlers

import (
	"github.com/dlomanov/mon/internal/apps/server/logger"
	"github.com/dlomanov/mon/internal/entities/metrics"
	"github.com/dlomanov/mon/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
)

func Get(db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawType := chi.URLParam(r, "type")
		mtype, ok := metrics.ParseMetricType(rawType)
		if !ok {
			http.NotFound(w, r)
			return
		}

		rawName := chi.URLParam(r, "name")
		if rawName == "" {
			http.NotFound(w, r)
			return
		}

		key := mtype.CreateKey(rawName)
		value, ok := db.Get(key)
		if !ok {
			http.NotFound(w, r)
			return
		}

		_, err := w.Write([]byte(value))
		if err != nil {
			logger.Log.Error("error occurred", zap.String("error", err.Error()))
		}
	}
}
