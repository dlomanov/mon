package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

func init() {
	var _ http.ResponseWriter = (*compressWriter)(nil)
	var _ io.Closer = (*compressWriter)(nil)
}

func newCompressWriter(w http.ResponseWriter, allowedTypes map[string]struct{}) *compressWriter {
	return &compressWriter{
		w:            w,
		cw:           gzip.NewWriter(w),
		contentTypes: allowedTypes,
		wroteHeader:  false,
		compressable: false,
	}
}

type compressWriter struct {
	w            http.ResponseWriter
	cw           io.WriteCloser
	contentTypes map[string]struct{}
	wroteHeader  bool
	compressable bool
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) WriteHeader(statusCode int) {
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Write(p []byte) (n int, err error) {
	c.writeHeader()
	return c.writer().Write(p)
}

func (c *compressWriter) Close() error {
	if !c.compressable {
		return nil
	}

	return c.cw.Close()
}

func (c *compressWriter) writeHeader() {
	if c.wroteHeader {
		return
	}
	c.wroteHeader = true

	c.compressable = c.isCompressable()
	if !c.compressable {
		return
	}

	c.w.Header().Set("Content-Encoding", compAlgo)
}

func (c *compressWriter) isCompressable() bool {
	contentType := c.Header().Get("Content-Type")
	if index := strings.Index(contentType, ";"); index >= 0 {
		contentType = contentType[0:index]
	}

	if _, ok := c.contentTypes[contentType]; ok {
		return true
	}

	return false
}

func (c *compressWriter) writer() io.Writer {
	if c.compressable {
		return c.cw
	}

	return c.w
}
