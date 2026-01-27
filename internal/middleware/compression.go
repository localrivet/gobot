package middleware

import (
	"compress/gzip"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

// Gzip wraps an HTTP handler with gzip compression
func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip compression if client doesn't support it
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Skip compression for WebSocket requests - they need http.Hijacker interface
		if strings.HasPrefix(r.URL.Path, "/ws") || r.Header.Get("Upgrade") == "websocket" {
			next.ServeHTTP(w, r)
			return
		}

		// Skip compression for API routes - they're proxied and small JSON
		if strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		// Skip compression for already compressed content types
		ext := filepath.Ext(r.URL.Path)
		if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" ||
			ext == ".webp" || ext == ".mp4" || ext == ".webm" || ext == ".pdf" {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length") // Length will change after compression

		gz := gzip.NewWriter(w)
		defer gz.Close()

		gzw := &gzipResponseWriter{Writer: gz, ResponseWriter: w}
		next.ServeHTTP(gzw, r)
	})
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
	statusCode int
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// ContentType sets correct MIME types based on file extension
func ContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ext := filepath.Ext(r.URL.Path)
		if ext != "" {
			if contentType := mime.TypeByExtension(ext); contentType != "" {
				w.Header().Set("Content-Type", contentType)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// CacheControl adds cache headers for static assets
func CacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Never cache API responses
		if strings.HasPrefix(path, "/api/") {
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
			next.ServeHTTP(w, r)
			return
		}

		// Immutable cache for hashed assets (SvelteKit _app/immutable/)
		if strings.Contains(path, "/_app/immutable/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else if strings.HasSuffix(path, ".html") || path == "/" {
			// No cache for HTML to ensure fresh content
			w.Header().Set("Cache-Control", "no-cache, must-revalidate")
		} else {
			// 24 hour cache for all other static assets
			w.Header().Set("Cache-Control", "public, max-age=86400")
		}

		next.ServeHTTP(w, r)
	})
}
