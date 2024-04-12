package middlewares

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Logger is a middleware that logs incoming HTTP requests. It captures details such as
// the request URI, method, elapsed time, response status code, and response size. These
// details are logged using the provided zap.Logger instance, which can be configured
// to output logs in various formats and destinations.
//
// The middleware wraps the provided HTTP handler with a custom response writer that
// captures the response status code and size. This allows the middleware to log these
// details after the handler has processed the request.
//
// The Logger middleware is useful for monitoring and debugging web server requests,
// providing insights into request handling performance and potential issues.
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
