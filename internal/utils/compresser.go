package utils

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
	w             http.ResponseWriter
	gw            *gzip.Writer
	headerWritten bool
	NeedCompress  bool
}

func NewCompressedWriter(w http.ResponseWriter) *compressedWriter {
	return &compressedWriter{
		w:  w,
		gw: gzip.NewWriter(w),
	}
}

func (c *compressedWriter) Write(data []byte) (int, error) {
	if !c.headerWritten {
		c.WriteHeader(http.StatusOK) // Default to 200 OK if no status is set.
	}
	if c.NeedCompress {
		return c.gw.Write(data)
	}
	return c.w.Write(data)
}

func (c *compressedWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressedWriter) WriteHeader(statusCode int) {
	if c.headerWritten {
		return
	}
	c.headerWritten = true
	contentType := c.Header().Get("Content-Type")
	if statusCode < 300 && (strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html")) {
		c.Header().Set("Content-Encoding", "gzip")
		c.NeedCompress = true
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressedWriter) Close() error {
	if c.NeedCompress {
		return c.gw.Close()
	}
	return nil
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func NewCompressReader(r io.ReadCloser) (*compressReader, error) {
	if r == nil {
		return nil, fmt.Errorf("request body is nil")
	}

	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c *compressReader) Read(p []byte) (int, error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.zr.Close(); err != nil {
		return err
	}
	return c.r.Close()
}

func CompressGzip(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, err := w.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to compress data: %v", err)
	}
	if err = w.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize compression: %v", err)
	}
	return b.Bytes(), nil
}

func DecompressGzip(body *bytes.Buffer) ([]byte, error) {
	gzipReader, err := gzip.NewReader(body)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzipReader.Close()

	var b bytes.Buffer
	_, err = io.Copy(&b, gzipReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %v", err)
	}

	return b.Bytes(), nil
}
