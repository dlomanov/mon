package middlewares

import (
	"net/http"
	"strings"
)

const compAlgo = "gzip"

var allowedTypes = map[string]struct{}{
	"application/json": {},
	"text/html":        {},
}

func Compressor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		if strings.Contains(acceptEncoding, compAlgo) {
			cw := newCompressWriter(w, allowedTypes)
			defer func(cw *compressWriter) { _ = cw.Close() }(cw)
			ow = cw
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		if strings.Contains(contentEncoding, compAlgo) {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer func(cr *compressReader) { _ = cr.Close() }(cr)
			r.Body = cr

		} else if contentEncoding != "" { // unsupported algo
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}

		next.ServeHTTP(ow, r)
	})
}
