package middlewares

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func Logger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrapper := &writer{
				ResponseWriter: w,
				data: &writerData{
					responseStatus: 200,
					responseSize:   0,
				},
			}

			start := time.Now()
			next.ServeHTTP(wrapper, r)

			logger.Info("incoming HTTP request",
				zap.String("URI", r.URL.Path),
				zap.String("method", r.Method),
				zap.Duration("elapsed_time", time.Since(start)),
				zap.Int("response_status_code", wrapper.data.responseStatus),
				zap.Int("response_size", wrapper.data.responseSize),
			)
		})
	}
}

type writer struct {
	http.ResponseWriter
	data *writerData
}

func (w *writer) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.data.responseStatus = statusCode
}

func (w *writer) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.data.responseSize += n
	return n, err
}

type writerData struct {
	responseStatus int
	responseSize   int
}
