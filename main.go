package main

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Embed static files into the binary
// The //go:embed directive tells the Go compiler to include these files
// at compile time. The result is a single binary with no external dependencies.
//
//go:embed static/*
var staticFiles embed.FS

// runHealthcheck checks if the server is responding and exits with appropriate code
// This is called when the binary is run with "healthcheck" argument
// Used by Docker HEALTHCHECK to verify the container is healthy
func runHealthcheck() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	resp, err := http.Get("http://localhost:" + port + "/health")
	if err != nil {
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		os.Exit(0)
	}
	os.Exit(1)
}

func main() {
	// Healthcheck mode: if run with "healthcheck" arg, just check if server responds
	// Example: ./demo-app healthcheck
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		runHealthcheck()
		return
	}

	// Configure structured JSON logging
	// All log output will be JSON for easy parsing by log aggregators
	//
	// If LOG_WEBHOOK_URL is set, logs are also POSTed to that URL.
	// This enables shipping logs to Splunk, Loki, or any HTTP endpoint
	// without requiring a sidecar or external agent.
	jsonHandler := slog.NewJSONHandler(os.Stdout, nil)

	webhookURL := os.Getenv("LOG_WEBHOOK_URL")
	webhookToken := os.Getenv("LOG_WEBHOOK_TOKEN")

	var handler slog.Handler
	if webhookURL != "" {
		// Wrap the JSON handler with webhook functionality
		handler = newWebhookHandler(jsonHandler, webhookURL, webhookToken)
	} else {
		// No webhook, just use the JSON handler directly
		handler = jsonHandler
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Log webhook status after logger is configured
	if webhookURL != "" {
		slog.Info("log webhook enabled", "url", webhookURL)
	}

	// Get configuration from environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = ":memory:"
	}

	// Initialize database
	// initStore is defined in store.go
	// db is a package-level variable in store.go
	var err error
	db, err = initStore(dbPath)
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize the sequence for auto-incrementing item IDs
	// The "100" is the bandwidth — it pre-allocates 100 IDs at a time for performance
	// itemSeq is a package-level variable in store.go
	itemSeq, err = db.GetSequence([]byte("seq:items"), 100)
	if err != nil {
		slog.Error("failed to initialize item sequence", "error", err)
		os.Exit(1)
	}
	defer itemSeq.Release()

	// Log database mode
	mode := "in-memory"
	if dbPath != "" && dbPath != ":memory:" {
		mode = "file"
	}
	slog.Info("database initialized", "path", dbPath, "mode", mode, "engine", "badger")

	// ==========================================================================
	// Route Registration
	// ==========================================================================
	//
	// Handlers are defined in handlers.go
	// loggingMiddleware is defined in middleware.go
	// All are accessible because they're in the same package (package main)

	// Health endpoint (for load balancers, Docker healthcheck)
	http.HandleFunc("/health", loggingMiddleware(healthHandler))

	// Items API (CRUD)
	http.HandleFunc("/api/items", loggingMiddleware(itemsHandler))
	http.HandleFunc("/api/items/", loggingMiddleware(itemsHandler)) // trailing slash catches /api/items/:id

	// Display panel API (arbitrary JSON storage)
	http.HandleFunc("/api/display", loggingMiddleware(displayHandler))

	// System info API (hostname, IPs, env vars)
	http.HandleFunc("/api/system", loggingMiddleware(systemHandler))

	// Prometheus metrics endpoint
	// No logging middleware — would be too noisy from Prometheus scraping every 15s
	http.Handle("/metrics", promhttp.Handler())

	// ==========================================================================
	// Static File Serving
	// ==========================================================================

	// Serve embedded static files (HTML, CSS, JS)
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		slog.Error("failed to create static file system", "error", err)
		os.Exit(1)
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Redirect root to dashboard
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/static/index.html", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})

	// ==========================================================================
	// Start Server
	// ==========================================================================

	slog.Info("server starting", "port", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}
