package handlers

import (
	"encoding/json"
	"github.com/dlomanov/mon/internal/apps/apimodels"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

func GetByParams(logger *zap.Logger, stg Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := apimodels.MetricKey{
			Name: chi.URLParam(r, "name"),
			Type: chi.URLParam(r, "type"),
		}

		entityKey, err := apimodels.MapToEntityKey(key)
		if err != nil {
			logger.Debug("invalid request body", zap.Error(err))
			http.NotFound(w, r)
			return
		}

		entity, ok := stg.Get(entityKey)
		if !ok {
			http.NotFound(w, r)
			return
		}

		_, err = w.Write([]byte(entity.StringValue()))
		if err != nil {
			logger.Error("error occurred during response writing", zap.Error(err))
		}
	}
}

func GetByJSON(logger *zap.Logger, db Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h := r.Header.Get(HeaderContentType); !strings.HasPrefix(h, "application/json") {
			logger.Debug("invalid content-type", zap.String(HeaderContentType, h))
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		var key apimodels.MetricKey
		err := json.NewDecoder(r.Body).Decode(&key)
		if err != nil {
			logger.Debug("cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		entityKey, err := apimodels.MapToEntityKey(key)
		if err != nil {
			logger.Debug("invalid request body", zap.Error(err))
			http.NotFound(w, r)
			return
		}

		entity, ok := db.Get(entityKey)
		if !ok {
			http.NotFound(w, r)
			return
		}

		metrics := apimodels.MapToModel(entity)
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(metrics)
		if err != nil {
			logger.Error("error occurred", zap.Error(err))
		}
	}
}
