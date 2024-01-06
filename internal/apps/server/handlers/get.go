package handlers

import (
	"encoding/json"
	"github.com/dlomanov/mon/internal/apps/apimodels"
	"github.com/dlomanov/mon/internal/apps/server/logger"
	"github.com/dlomanov/mon/internal/entities"
	"github.com/dlomanov/mon/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

func GetByParams(db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := apimodels.MetricKey{
			Name: chi.URLParam(r, "name"),
			Type: chi.URLParam(r, "type"),
		}

		entityKey, err := apimodels.MapToEntityKey(key)
		if err != nil {
			logger.Log.Debug("invalid request body", zap.Error(err))
			http.NotFound(w, r)
			return
		}

		value, ok := db.Get(entityKey.String())
		if !ok {
			http.NotFound(w, r)
			return
		}

		_, err = w.Write([]byte(value))
		if err != nil {
			logger.Log.Error("error occurred during response writing", zap.Error(err))
		}
	}
}

func GetByJSON(db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h := r.Header.Get(HeaderContentType); !strings.HasPrefix(h, "application/json") {
			logger.Log.Debug("invalid content-type", zap.String(HeaderContentType, h))
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		var key apimodels.MetricKey
		err := json.NewDecoder(r.Body).Decode(&key)
		if err != nil {
			logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		entityKey, err := apimodels.MapToEntityKey(key)
		if err != nil {
			logger.Log.Debug("invalid request body", zap.Error(err))
			http.NotFound(w, r)
			return
		}

		value, ok := db.Get(entityKey.String())
		if !ok {
			http.NotFound(w, r)
			return
		}

		entity := entities.Metric{MetricsKey: entityKey}
		entity, err = entity.CloneWith(value)
		if err != nil {
			logger.Log.Error("error occurred", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		metrics := apimodels.MapToModel(entity)
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(metrics)
		if err != nil {
			logger.Log.Error("error occurred", zap.Error(err))
		}
	}
}
