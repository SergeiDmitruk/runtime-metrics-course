package compress

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var cwInstance compressedWriter
var crInstance compressReader

type compressedWriter struct {
	w  http.ResponseWriter
	gw *gzip.Writer
}

func NewCompressedWriter(w http.ResponseWriter) *compressedWriter {
	return &compressedWriter{
		w:  w,
		gw: gzip.NewWriter(w),
	}
}

func (c *compressedWriter) Write(data []byte) (int, error) {
	ct := c.w.Header().Get("Content-type")
	if strings.Contains(ct, "text/html") || strings.Contains(ct, "application/json") {
		return c.gw.Write(data)
	}
	return c.w.Write(data)
}
func (c *compressedWriter) Header() http.Header {
	return c.w.Header()
}
func (c *compressedWriter) Close() error {
	return c.gw.Close()
}

func (c *compressedWriter) WriteHeader(statusCode int) {
	ct := c.w.Header().Get("Content-type")
	if statusCode < 300 && (strings.Contains(ct, "text/html") || strings.Contains(ct, "application/json")) {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func NewCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
func CompressGzip(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, err := w.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}
	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %v", err)
	}
	return b.Bytes(), nil
}

func DecompressGzip(r io.ReadCloser) ([]byte, error) {

	r, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed decompress data: %v", err)
	}
	defer r.Close()
	var b bytes.Buffer
	_, err = b.ReadFrom(r)
	if err != nil {
		return nil, fmt.Errorf("failed decompress data: %v", err)
	}

	return b.Bytes(), nil
}
