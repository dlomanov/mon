package handlers

import (
	"context"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const timeout = 5 * time.Second

func (c *Container) PingDB() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timeoutCtx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		if c.DB == nil {
			c.Logger.Debug("DB is not configured")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := c.DB.PingContext(timeoutCtx); err != nil {
			c.Logger.Error("failed ping to DB", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
