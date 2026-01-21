package main

import (
	"encoding/json"
	"time"

	badger "github.com/dgraph-io/badger/v4"
)

// Key prefix for items in BadgerDB
// All item keys look like: "item:1", "item:2", etc.
// Using a prefix lets us iterate over just items (not other data we might store)
const itemKeyPrefix = "item:"

// Package-level database connection
// Handlers need access to this to read/write data
var db *badger.DB

// Sequence for auto-incrementing item IDs
// BadgerDB sequences are atomic and safe for concurrent access
var itemSeq *badger.Sequence

// Package-level display data (in-memory, transient)
// This is NOT stored in BadgerDB — it resets when the app restarts
// json.RawMessage holds arbitrary JSON without parsing it
var displayData json.RawMessage

// Item represents a generic item in the database
// The struct tags (json:"...") control how Go marshals/unmarshals JSON
// omitempty means the field is excluded from JSON if it's empty
type Item struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// initStore opens the BadgerDB database
// dbPath can be:
//   - empty string or ":memory:" for in-memory (ephemeral)
//   - a directory path for persistent storage
//
// Returns the opened database or an error
// The caller (main) is responsible for closing it with defer db.Close()
func initStore(dbPath string) (*badger.DB, error) {
	var opts badger.Options

	// Determine if we're using in-memory or file-based storage
	if dbPath == "" || dbPath == ":memory:" {
		// In-memory mode: fast, ephemeral, data lost on restart
		opts = badger.DefaultOptions("").WithInMemory(true)
	} else {
		// File-based mode: persistent, data survives restarts
		opts = badger.DefaultOptions(dbPath)
	}

	// Reduce logging noise from BadgerDB (it's verbose by default)
	opts = opts.WithLoggingLevel(badger.WARNING)

	// Open the database
	database, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return database, nil
}
