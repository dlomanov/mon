package middlewares

import (
	"compress/gzip"
	"io"
)

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	d, err := gzip.NewReader(r)
	return &compressReader{
		reader:       r,
		decompressor: d,
	}, err
}

type compressReader struct {
	reader       io.ReadCloser
	decompressor io.ReadCloser
}

func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.decompressor.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.reader.Close(); err != nil {
		return err
	}
	return c.decompressor.Close()
}
