package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

var skipPaths = []string{"/metrics", "/health", "/live", "/ready"}

type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter   *gzip.Writer
	wroteHeader bool
	statusCode  int
	size        int64
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.statusCode = code
		w.wroteHeader = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.gzipWriter.Write(data)
	w.size += int64(n)
	return n, err
}

func (w *gzipResponseWriter) Close() error {
	return w.gzipWriter.Close()
}

func (w *gzipResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
	w.gzipWriter.Flush()
}

func GzipHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !shouldCompress(r) {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		gzw := &gzipResponseWriter{
			ResponseWriter: w,
			gzipWriter:     gz,
		}

		next.ServeHTTP(gzw, r)
	})
}

func shouldCompress(r *http.Request) bool {
	ae := r.Header.Get("Accept-Encoding")
	if !strings.Contains(ae, "gzip") {
		return false
	}

	for _, skip := range skipPaths {
		if strings.HasPrefix(r.URL.Path, skip) {
			return false
		}
	}

	if r.Method == http.MethodHead || r.Method == http.MethodOptions {
		return false
	}

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "text/event-stream") {
		return false
	}

	return true
}
