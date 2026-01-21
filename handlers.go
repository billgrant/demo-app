package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	badger "github.com/dgraph-io/badger/v4"
)

// =============================================================================
// Health Endpoint
// =============================================================================

// healthHandler responds with a JSON health status
// Used by Docker HEALTHCHECK and load balancers to verify the app is running
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// =============================================================================
// Items Endpoints (CRUD)
// =============================================================================

// itemsHandler routes /api/items requests based on method and path
// This is a "sub-router" pattern — one handler that dispatches to others
// Python equivalent: a Flask blueprint with multiple routes
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

// listItems returns all items from the database
func listItems(w http.ResponseWriter, r *http.Request) {
	items := []Item{}

	// db.View() starts a read-only transaction
	// This is safe for concurrent access — multiple readers can run simultaneously
	err := db.View(func(txn *badger.Txn) error {
		// Create an iterator with default options
		opts := badger.DefaultIteratorOptions
		// PrefetchValues = true means we want the values, not just keys
		opts.PrefetchValues = true

		it := txn.NewIterator(opts)
		defer it.Close()

		// Seek to the first key with our prefix, then iterate while prefix matches
		prefix := []byte(itemKeyPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			// Get the value (the JSON blob)
			err := item.Value(func(val []byte) error {
				var i Item
				if err := json.Unmarshal(val, &i); err != nil {
					slog.Error("failed to unmarshal item", "error", err)
					return nil // Skip malformed items, don't fail the whole list
				}
				items = append(items, i)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		slog.Error("failed to list items", "error", err)
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(items)
}

// createItem creates a new item in the database
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

	// Get next ID from the sequence
	// This is atomic and safe for concurrent access
	id, err := itemSeq.Next()
	if err != nil {
		slog.Error("failed to get next item ID", "error", err)
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}

	// Create the item
	item := Item{
		ID:          int64(id),
		Name:        input.Name,
		Description: input.Description,
		CreatedAt:   time.Now().UTC(),
	}

	// Serialize to JSON
	value, err := json.Marshal(item)
	if err != nil {
		slog.Error("failed to marshal item", "error", err)
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}

	// Build the key: "item:1", "item:2", etc.
	key := []byte(fmt.Sprintf("%s%d", itemKeyPrefix, id))

	// db.Update() starts a read-write transaction
	// Multiple Update transactions are serialized, but this is fast for K/V operations
	err = db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
	if err != nil {
		slog.Error("failed to insert item", "error", err)
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}

	// Update Prometheus metrics (defined in metrics.go)
	itemsTotal.Inc()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

// getItem returns a single item by ID
func getItem(w http.ResponseWriter, r *http.Request, id int64) {
	key := []byte(fmt.Sprintf("%s%d", itemKeyPrefix, id))
	var item Item

	err := db.View(func(txn *badger.Txn) error {
		dbItem, err := txn.Get(key)
		if err != nil {
			return err // Will be badger.ErrKeyNotFound if not exists
		}

		return dbItem.Value(func(val []byte) error {
			return json.Unmarshal(val, &item)
		})
	})

	if err == badger.ErrKeyNotFound {
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

	key := []byte(fmt.Sprintf("%s%d", itemKeyPrefix, id))
	var item Item

	// Update is a read-modify-write operation, all in one transaction
	err := db.Update(func(txn *badger.Txn) error {
		// First, read the existing item
		dbItem, err := txn.Get(key)
		if err != nil {
			return err // badger.ErrKeyNotFound if doesn't exist
		}

		// Get current value and unmarshal
		err = dbItem.Value(func(val []byte) error {
			return json.Unmarshal(val, &item)
		})
		if err != nil {
			return err
		}

		// Update fields (preserve CreatedAt and ID)
		item.Name = input.Name
		item.Description = input.Description

		// Marshal and save
		value, err := json.Marshal(item)
		if err != nil {
			return err
		}

		return txn.Set(key, value)
	})

	if err == badger.ErrKeyNotFound {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("failed to update item", "error", err)
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(item)
}

// deleteItem removes an item by ID
func deleteItem(w http.ResponseWriter, r *http.Request, id int64) {
	key := []byte(fmt.Sprintf("%s%d", itemKeyPrefix, id))

	// First check if the item exists (for proper 404 handling)
	err := db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		return err
	})

	if err == badger.ErrKeyNotFound {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("failed to check item existence", "error", err)
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}

	// Item exists, delete it
	err = db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
	if err != nil {
		slog.Error("failed to delete item", "error", err)
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}

	// Update Prometheus metrics (defined in metrics.go)
	itemsTotal.Dec()

	w.WriteHeader(http.StatusNoContent)
}

// =============================================================================
// Display Endpoints
// =============================================================================

// displayHandler handles GET/POST for the display panel
// GET returns current data, POST replaces it with new data
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
// The data is stored in memory (displayData variable from store.go)
// and is lost when the app restarts
func setDisplay(w http.ResponseWriter, r *http.Request) {
	// Read the raw JSON body
	var data json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	// Store it (package-level variable from store.go)
	displayData = data

	// Update Prometheus metrics (defined in metrics.go)
	displayUpdatesTotal.Inc()

	// Return what we stored
	w.WriteHeader(http.StatusCreated)
	w.Write(displayData)
}

// =============================================================================
// System Endpoint
// =============================================================================

// systemHandler returns system information (GET only)
// Used to verify deployment location, container info, etc.
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

	// Get request headers (useful for debugging proxy chains, auth, etc.)
	// r.Header is map[string][]string — headers can have multiple values
	headers := getRequestHeaders(r)

	// Get client info from the request (demo-friendly, shows "who's hitting the app")
	// r.RemoteAddr is the client's IP:port
	// r.UserAgent() is a convenience method for the User-Agent header
	clientIP := r.RemoteAddr
	userAgent := r.UserAgent()

	response := map[string]interface{}{
		"hostname":    hostname,
		"ips":         ips,
		"environment": envVars,
		"headers":     headers,
		"client_ip":   clientIP,
		"user_agent":  userAgent,
	}

	json.NewEncoder(w).Encode(response)
}

// getIPAddresses returns all non-loopback IPv4 addresses
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
// Only returns values for explicitly allowed variable names
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

// getRequestHeaders returns HTTP headers from the incoming request
// Useful for debugging:
//   - Proxy chains: X-Forwarded-For, X-Real-IP
//   - Load balancers: X-Forwarded-Proto, X-Forwarded-Host
//   - Auth: Authorization (shows if present, not the value for security)
//   - Client info: User-Agent, Accept, Accept-Language
func getRequestHeaders(r *http.Request) map[string]string {
	headers := make(map[string]string)

	for name, values := range r.Header {
		if len(values) == 1 {
			// Most headers have single values
			headers[name] = values[0]
		} else {
			// Multiple values: join with comma (standard HTTP format)
			headers[name] = strings.Join(values, ", ")
		}
	}

	return headers
}
