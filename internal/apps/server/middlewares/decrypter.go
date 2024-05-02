package middlewares

import (
	"bytes"
	"io"
	"net/http"

	"github.com/dlomanov/mon/internal/services/encrypt"
	"go.uber.org/zap"
)

func Decrypter(logger *zap.Logger, dec *encrypt.Decryptor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, ok := r.Header["Encryption"]
			if dec == nil || !ok {
				next.ServeHTTP(w, r)
				return
			}

			encBody, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Error("failed to read body", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body.Close()
			body, err := dec.Decrypt(encBody)
			if err != nil {
				logger.Error("failed to decrypt body", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))

			next.ServeHTTP(w, r)
		})
	}
}
