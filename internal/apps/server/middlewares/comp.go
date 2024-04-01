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

// Compressor is a middleware that compresses HTTP responses using gzip if the client
// supports it. It checks the "Accept-Encoding" header of the request to determine if
// the client can accept gzip-encoded responses. If so, it wraps the response writer
// with a gzip writer. It also handles decompression of the request body if it's
// gzip-encoded, by replacing the request body with a decompressed reader.
//
// The middleware supports only "application/json" and "text/html" content types for
// compression. It returns a new HTTP handler that wraps the provided handler with
// compression functionality.
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
