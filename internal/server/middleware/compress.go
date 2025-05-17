package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

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

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

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
func CompressResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ow := res
		acceptEncoding := req.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := newCompressWriter(res)
			ow = cw
			defer cw.Close()
		}
		next.ServeHTTP(ow, req)
	})
}
