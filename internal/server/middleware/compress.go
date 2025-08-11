package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// DecompressRequest returns middleware that decompresses gzip-encoded request bodies.
func DecompressRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		var reader io.Reader
		if req.Header.Get(`Content-Encoding`) == `gzip` {
			gz, err := gzip.NewReader(req.Body)
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			reader = gz
			defer gz.Close()
		} else {
			reader = req.Body
		}

		body, err := io.ReadAll(reader)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		req.Body = io.NopCloser(bytes.NewReader(body))
		next.ServeHTTP(res, req)
	})
}

// compressWriter wraps an http.ResponseWriter to provide gzip compression.
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// newCompressWriter creates a new compressWriter that wraps the provided ResponseWriter.
func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// CompressResponse returns middleware that compresses response bodies with gzip when the client supports it.
func CompressResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ow := res
		acceptEncoding := req.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := newCompressWriter(res)
			ow = cw
			defer cw.Close()
			ow.Header().Add("Content-Encoding", "gzip")
		}
		next.ServeHTTP(ow, req)
	})
}
