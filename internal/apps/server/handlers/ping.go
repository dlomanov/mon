package handlers

import (
	"context"
	"database/sql"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const timeout = 5 * time.Second

func PingDB(ctx context.Context, logger *zap.Logger, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		if err := db.PingContext(timeoutCtx); err != nil {
			logger.Error("failed ping to DB", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
