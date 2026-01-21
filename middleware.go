package main

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// responseRecorder wraps http.ResponseWriter to capture the status code
// Go's ResponseWriter doesn't expose the status after WriteHeader is called,
// so we wrap it to intercept and store the value
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before passing it through
func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// loggingMiddleware wraps a handler to log every request and record Prometheus metrics
// This is the "middleware pattern" — a function that takes a handler and returns a new handler
// Python equivalent: a decorator that wraps a Flask route
func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the ResponseWriter to capture status code
		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     200, // default if WriteHeader isn't called
		}

		// Call the actual handler
		next(recorder, r)

		// Calculate duration
		duration := time.Since(start)

		// Normalize path for metrics to avoid high cardinality
		// /api/items/123 -> /api/items/:id (prevents explosion of metric series)
		metricPath := normalizePath(r.URL.Path)

		// Log the request (original path for debugging)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.statusCode,
			"latency_ms", duration.Milliseconds(),
			"client_ip", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)

		// Record Prometheus metrics
		// These variables are defined in metrics.go but accessible here (same package)
		httpRequestsTotal.WithLabelValues(
			r.Method,
			metricPath,
			strconv.Itoa(recorder.statusCode),
		).Inc()

		httpRequestDuration.WithLabelValues(
			r.Method,
			metricPath,
		).Observe(duration.Seconds())
	}
}

// normalizePath replaces dynamic path segments with placeholders
// This prevents high cardinality in Prometheus metrics
// Example: /api/items/123 -> /api/items/:id
//
// Why this matters: If we used the raw path, we'd create a new metric series
// for every unique item ID. With millions of items, that's millions of series,
// which would overwhelm Prometheus.
func normalizePath(path string) string {
	// Handle /api/items/:id pattern
	if strings.HasPrefix(path, "/api/items/") {
		parts := strings.Split(path, "/")
		if len(parts) == 4 && parts[3] != "" {
			// /api/items/123 -> 4 parts: ["", "api", "items", "123"]
			return "/api/items/:id"
		}
	}
	return path
}
