package middlewares

import (
	"bytes"
	"errors"
	"github.com/dlomanov/mon/internal/infra/services/hashing"
	"io"
	"net/http"

	"github.com/dlomanov/mon/internal/apps/server/container"
	"go.uber.org/zap"
)

// Hash is a middleware that validates and optionally computes a hash for HTTP requests
// and responses.
//
// It checks for a predefined header (e.g., "X-Hash") in the request and
// compares it to a computed hash based on the request body and a secret key. If the
// header is present and the hashes match, the request is considered valid. Otherwise,
// the middleware may respond with an error status code.
//
// The middleware also computes a hash for the response body and sets it in the response header.
// This allows the client to verify the integrity of
// the response.
//
// The middleware function returns a new HTTP handler that wraps the provided handler with
// hash validation and computation functionality.
func Hash(c *container.Container) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// well... there's nothing we can do
			if c.Config.Key == "" {
				next.ServeHTTP(w, r)
				return
			}

			if !validate(c, w, r) {
				return
			}

			bw := newBufferedWriter(w)
			defer flush(bw, c)

			next.ServeHTTP(bw, r)
			setHashHeader(c, bw)
		})
	}
}

func validate(c *container.Container, w http.ResponseWriter, r *http.Request) (ok bool) {
	headerValue := r.Header.Get(hashing.HeaderHash)
	if headerValue == "" {
		return true
	}

	value, err := io.ReadAll(r.Body)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			c.Logger.Error("request body reading failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return false
		}
	}
	r.Body = io.NopCloser(bytes.NewReader(value))

	hash := hashing.ComputeBase64URLHash(c.Config.Key, value)
	if hash != headerValue {
		c.Logger.Debug("invalid hash",
			zap.String("client_hash", headerValue),
			zap.String("server_hash", hash))
		w.WriteHeader(http.StatusBadRequest)
		return false
	}

	return true
}

func setHashHeader(c *container.Container, bw *bufferedWriter) {
	if bw.Len() == 0 {
		return
	}

	hash := hashing.ComputeBase64URLHash(c.Config.Key, bw.buffer.Bytes())
	bw.Header().Set(hashing.HeaderHash, hash)
}

func flush(bw *bufferedWriter, c *container.Container) {
	if err := bw.Flush(); err != nil {
		c.Logger.Error("buffer flushing failed", zap.Error(err))
		bw.buffer.Reset()
	}
}

var _ http.ResponseWriter = (*bufferedWriter)(nil)

type bufferedWriter struct {
	http.ResponseWriter
	buffer *bytes.Buffer
}

func newBufferedWriter(w http.ResponseWriter) *bufferedWriter {
	return &bufferedWriter{
		ResponseWriter: w,
		buffer:         new(bytes.Buffer),
	}
}

func (bw *bufferedWriter) Write(p []byte) (int, error) {
	return bw.buffer.Write(p)
}

func (bw *bufferedWriter) Len() int {
	return bw.buffer.Len()
}

func (bw *bufferedWriter) Flush() error {
	data := bw.buffer.Bytes()
	_, err := bw.ResponseWriter.Write(data)
	if err != nil {
		return err
	}
	bw.buffer.Reset()
	return nil
}
