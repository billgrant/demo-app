package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// webhookHandler wraps another slog.Handler and optionally sends logs to a webhook.
//
// This implements the slog.Handler interface, which requires 4 methods:
//   - Enabled()   — should this log level be logged?
//   - Handle()    — process a log record (this is where the magic happens)
//   - WithAttrs() — create a new handler with additional attributes
//   - WithGroup() — create a new handler with a group prefix
//
// The struct holds DATA, the methods define BEHAVIOR.
type webhookHandler struct {
	underlying slog.Handler // the wrapped handler (JSONHandler for stdout)
	webhookURL string       // where to POST logs (empty = disabled)
	token      string       // optional auth token
	client     *http.Client // reusable HTTP client
}

// newWebhookHandler creates a handler that writes to stdout AND posts to a webhook.
//
// Parameters:
//   - underlying: the handler that writes to stdout (typically JSONHandler)
//   - webhookURL: URL to POST logs to (empty string disables webhook)
//   - token: optional Authorization header value
//
// Returns a handler that satisfies slog.Handler interface.
func newWebhookHandler(underlying slog.Handler, webhookURL, token string) *webhookHandler {
	return &webhookHandler{
		underlying: underlying,
		webhookURL: webhookURL,
		token:      token,
		// Custom HTTP client with timeout — don't let slow webhooks hang forever
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// =============================================================================
// slog.Handler interface implementation
// =============================================================================

// Enabled reports whether the handler handles records at the given level.
// We delegate to the underlying handler — if it wouldn't log this level, neither do we.
func (w *webhookHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return w.underlying.Enabled(ctx, level)
}

// Handle processes a log record. This is called for every log statement.
//
// Our logic:
//  1. Always pass to underlying handler (writes to stdout)
//  2. If webhook is configured, also POST the log entry (async)
//
// The context parameter carries request-scoped data (deadlines, cancellation).
// We ignore it for the async POST since we want logs to ship even if the
// original request context is cancelled.
func (w *webhookHandler) Handle(ctx context.Context, record slog.Record) error {
	// Step 1: Always write to stdout via the underlying handler
	if err := w.underlying.Handle(ctx, record); err != nil {
		return err
	}

	// Step 2: If webhook is configured, POST asynchronously
	if w.webhookURL != "" {
		// Build the log entry as a map
		entry := w.buildLogEntry(record)

		// Launch goroutine — don't block the request waiting for webhook
		// This is "fire and forget" — we don't wait for the result
		go w.postToWebhook(entry)
	}

	return nil
}

// WithAttrs returns a new handler with additional attributes.
// This is called when you do: logger.With("key", "value")
//
// We need to wrap the underlying handler's WithAttrs result,
// keeping our webhook config intact.
func (w *webhookHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &webhookHandler{
		underlying: w.underlying.WithAttrs(attrs),
		webhookURL: w.webhookURL,
		token:      w.token,
		client:     w.client,
	}
}

// WithGroup returns a new handler with a group prefix.
// This is called when you do: logger.WithGroup("request")
// Then subsequent attrs become "request.key" instead of "key".
//
// Same pattern as WithAttrs — wrap the result, keep our config.
func (w *webhookHandler) WithGroup(name string) slog.Handler {
	return &webhookHandler{
		underlying: w.underlying.WithGroup(name),
		webhookURL: w.webhookURL,
		token:      w.token,
		client:     w.client,
	}
}

// =============================================================================
// Webhook logic
// =============================================================================

// buildLogEntry converts a slog.Record into a map for JSON serialization.
//
// slog.Record contains:
//   - Time: when the log was created
//   - Level: INFO, WARN, ERROR, etc.
//   - Message: the log message
//   - Attrs: key-value pairs added via slog.Info("msg", "key", "value")
func (w *webhookHandler) buildLogEntry(record slog.Record) map[string]any {
	entry := map[string]any{
		"time":  record.Time.Format(time.RFC3339),
		"level": record.Level.String(),
		"msg":   record.Message,
	}

	// Iterate over all attributes and add them to the entry
	// record.Attrs is a method that takes a callback — Go's iterator pattern
	record.Attrs(func(attr slog.Attr) bool {
		entry[attr.Key] = attr.Value.Any()
		return true // continue iterating
	})

	return entry
}

// postToWebhook sends a log entry to the configured webhook URL.
//
// This runs in a goroutine (async), so it:
//   - Doesn't block the HTTP request
//   - Doesn't return errors to the caller (just logs failures to stderr)
//   - Uses its own timeout (5 seconds) independent of request context
func (w *webhookHandler) postToWebhook(entry map[string]any) {
	// Serialize to JSON
	body, err := json.Marshal(entry)
	if err != nil {
		// Log to stderr — can't use slog here (would cause infinite loop!)
		// Using println as a simple fallback
		println("webhook: failed to marshal log entry:", err.Error())
		return
	}

	// Create the request
	req, err := http.NewRequest(http.MethodPost, w.webhookURL, bytes.NewReader(body))
	if err != nil {
		println("webhook: failed to create request:", err.Error())
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if w.token != "" {
		req.Header.Set("Authorization", w.token)
	}

	// Send the request
	resp, err := w.client.Do(req)
	if err != nil {
		println("webhook: failed to send:", err.Error())
		return
	}
	defer resp.Body.Close()

	// Check for non-2xx response
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		println("webhook: unexpected status:", resp.StatusCode)
	}
}
