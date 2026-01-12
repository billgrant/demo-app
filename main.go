package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite" // SQLite driver (registers itself with database/sql)
)

//go:embed static/*
var staticFiles embed.FS

// Package-level database connection (handlers need access)
var db *sql.DB

// Package-level display data (in-memory, transient)
var displayData json.RawMessage

// Item represents a generic item in the database
type Item struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

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

// initDB opens the SQLite database and creates tables
func initDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Verify connection works
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Create items table for Phase 2 CRUD
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// runHealthcheck checks if the server is responding and exits with appropriate code
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
	// Healthcheck mode: check if server is responding
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		runHealthcheck()
		return
	}

	// Configure structured JSON logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Get configuration from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = ":memory:"
	}

	// Initialize database (assigns to package-level var)
	var err error
	db, err = initDB(dbPath)
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("database initialized", "path", dbPath)

	// Register endpoints with logging middleware
	http.HandleFunc("/health", loggingMiddleware(healthHandler))
	http.HandleFunc("/api/items", loggingMiddleware(itemsHandler))
	http.HandleFunc("/api/items/", loggingMiddleware(itemsHandler)) // trailing slash catches /api/items/:id
	http.HandleFunc("/api/display", loggingMiddleware(displayHandler))
	http.HandleFunc("/api/system", loggingMiddleware(systemHandler))

	// Serve embedded static files
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

	// Start the server
	slog.Info("server starting", "port", port)
	err = http.ListenAndServe(":"+port, nil)
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

// itemsHandler routes /api/items requests based on method and path
func itemsHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path if present: /api/items/123 -> "123"
	path := strings.TrimPrefix(r.URL.Path, "/api/items")
	path = strings.TrimPrefix(path, "/")

	w.Header().Set("Content-Type", "application/json")

	// Route based on method and whether we have an ID
	if path == "" {
		// /api/items (no ID)
		switch r.Method {
		case http.MethodGet:
			listItems(w, r)
		case http.MethodPost:
			createItem(w, r)
		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	} else {
		// /api/items/:id
		id, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			getItem(w, r, id)
		case http.MethodPut:
			updateItem(w, r, id)
		case http.MethodDelete:
			deleteItem(w, r, id)
		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	}
}

// listItems returns all items
func listItems(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, description, created_at FROM items ORDER BY created_at DESC")
	if err != nil {
		slog.Error("failed to query items", "error", err)
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := []Item{}
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt); err != nil {
			slog.Error("failed to scan item", "error", err)
			continue
		}
		items = append(items, item)
	}

	json.NewEncoder(w).Encode(items)
}

// createItem creates a new item
func createItem(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if input.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO items (name, description) VALUES (?, ?)", input.Name, input.Description)
	if err != nil {
		slog.Error("failed to insert item", "error", err)
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()

	// Fetch the created item to return it
	var item Item
	err = db.QueryRow("SELECT id, name, description, created_at FROM items WHERE id = ?", id).
		Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt)
	if err != nil {
		slog.Error("failed to fetch created item", "error", err)
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

// getItem returns a single item by ID
func getItem(w http.ResponseWriter, r *http.Request, id int64) {
	var item Item
	err := db.QueryRow("SELECT id, name, description, created_at FROM items WHERE id = ?", id).
		Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("failed to fetch item", "error", err)
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(item)
}

// updateItem updates an existing item
func updateItem(w http.ResponseWriter, r *http.Request, id int64) {
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if input.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}

	result, err := db.Exec("UPDATE items SET name = ?, description = ? WHERE id = ?", input.Name, input.Description, id)
	if err != nil {
		slog.Error("failed to update item", "error", err)
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	// Fetch and return the updated item
	var item Item
	db.QueryRow("SELECT id, name, description, created_at FROM items WHERE id = ?", id).
		Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt)

	json.NewEncoder(w).Encode(item)
}

// deleteItem removes an item by ID
func deleteItem(w http.ResponseWriter, r *http.Request, id int64) {
	result, err := db.Exec("DELETE FROM items WHERE id = ?", id)
	if err != nil {
		slog.Error("failed to delete item", "error", err)
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// displayHandler handles GET/POST for the display panel
func displayHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		getDisplay(w, r)
	case http.MethodPost:
		setDisplay(w, r)
	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

// getDisplay returns the current display data
func getDisplay(w http.ResponseWriter, r *http.Request) {
	if displayData == nil {
		// Return empty object if nothing set
		w.Write([]byte("{}"))
		return
	}
	w.Write(displayData)
}

// setDisplay stores arbitrary JSON for display
func setDisplay(w http.ResponseWriter, r *http.Request) {
	// Read the raw JSON body
	var data json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	// Store it
	displayData = data

	// Return what we stored
	w.WriteHeader(http.StatusCreated)
	w.Write(displayData)
}

// systemHandler returns system information (GET only)
func systemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Get network interfaces and IPs
	ips := getIPAddresses()

	// Get selected environment variables (safe to expose)
	envVars := getFilteredEnvVars()

	response := map[string]interface{}{
		"hostname":    hostname,
		"ips":         ips,
		"environment": envVars,
	}

	json.NewEncoder(w).Encode(response)
}

// getIPAddresses returns all non-loopback IP addresses
func getIPAddresses() []string {
	var ips []string

	interfaces, err := net.Interfaces()
	if err != nil {
		return ips
	}

	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			// Extract IP from CIDR notation
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipnet.IP.To4() != nil { // IPv4 only for simplicity
					ips = append(ips, ipnet.IP.String())
				}
			}
		}
	}

	return ips
}

// getFilteredEnvVars returns environment variables safe to expose
func getFilteredEnvVars() map[string]string {
	// Allowlist of env vars to expose
	allowed := []string{
		"PORT",
		"DB_PATH",
		"HOSTNAME",      // Set by Docker/K8s
		"POD_NAME",      // Kubernetes
		"POD_NAMESPACE", // Kubernetes
		"NODE_NAME",     // Kubernetes
		"CONTAINER_ID",  // Docker
	}

	result := make(map[string]string)
	for _, key := range allowed {
		if val := os.Getenv(key); val != "" {
			result[key] = val
		}
	}
	return result
}
