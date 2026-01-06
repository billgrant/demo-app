package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// responseRecorder wraps http.ResponseWriter to capture the status code
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before passing it through
func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// loggingMiddleware wraps a handler to log every request
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

		// Log the request
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.statusCode,
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	}
}

func main() {
	// Configure structured JSON logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Get port from environment, default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Register the health endpoint with logging middleware
	http.HandleFunc("/health", loggingMiddleware(healthHandler))

	// Start the server
	slog.Info("server starting", "port", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}

// healthHandler responds with a JSON health status
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
