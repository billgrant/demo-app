// Simple HTTP server that receives webhook POSTs and prints them.
//
// Usage:
//   go run scripts/webhook-receiver/main.go
//   go run scripts/webhook-receiver/main.go -port 9999
//
// Then in another terminal:
//   LOG_WEBHOOK_URL="http://localhost:9999/logs" ./demo-app

package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	port := flag.String("port", "9999", "port to listen on")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		fmt.Printf("\n[%s] %s %s\n", time.Now().Format("15:04:05"), r.Method, r.URL.Path)
		fmt.Printf("Body: %s\n", string(body))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"received"}`))
	})

	fmt.Printf("Webhook receiver listening on port %s...\n", *port)
	fmt.Printf("Send logs to: http://localhost:%s/logs\n", *port)
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	http.ListenAndServe(":"+*port, nil)
}
