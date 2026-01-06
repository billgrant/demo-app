package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	// Get port from environment, default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Register the health endpoint
	http.HandleFunc("/health", healthHandler)

	// Start the server
	log.Printf("Starting server on port %s", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
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
