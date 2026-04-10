package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// wrappedWriter captures the status code written by handlers.
type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Logger logs every request: method, path, status, duration, request ID.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &wrappedWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(ww, r)

		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", GetRequestID(r.Context()),
		)
	})
}
