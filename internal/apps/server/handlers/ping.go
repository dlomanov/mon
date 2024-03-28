package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/dlomanov/mon/internal/apps/server/container"
	"go.uber.org/zap"
)

const timeout = 5 * time.Second

func PingDB(c *container.Container) http.HandlerFunc {
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
