package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestMain runs once before all tests in this file.
// It initializes the database so handlers have a working store.
// This is Go's way of doing "setup" for a test suite — like pytest fixtures.
func TestMain(m *testing.M) {
	// Initialize BadgerDB in-memory for tests
	var err error
	db, err = initStore(":memory:")
	if err != nil {
		panic("failed to init test database: " + err.Error())
	}
	defer db.Close()

	// Initialize the item ID sequence
	itemSeq, err = db.GetSequence([]byte("seq:items"), 100)
	if err != nil {
		panic("failed to init test sequence: " + err.Error())
	}
	defer itemSeq.Release()

	// Run all tests
	os.Exit(m.Run())
}

// resetDisplayData clears the display panel between tests
func resetDisplayData() {
	displayData = nil
}

// =============================================================================
// Health Endpoint Tests
// =============================================================================

func TestHealthHandler_ReturnsOK(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	healthHandler(rr, req)

	// Check status code
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	// Check response body has expected fields
	var result map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%s'", result["status"])
	}

	if _, ok := result["timestamp"]; !ok {
		t.Error("expected 'timestamp' field in response")
	}
}

// =============================================================================
// Items Endpoint Tests
// =============================================================================

func TestItems_CreateAndList(t *testing.T) {
	// Create an item
	body := bytes.NewBufferString(`{"name":"Test Item","description":"A test"}`)
	req := httptest.NewRequest("POST", "/api/items", body)
	rr := httptest.NewRecorder()

	itemsHandler(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("create: expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}

	// Parse the created item to get its ID
	var created Item
	if err := json.Unmarshal(rr.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to parse created item: %v", err)
	}

	if created.Name != "Test Item" {
		t.Errorf("expected name 'Test Item', got '%s'", created.Name)
	}
	if created.Description != "A test" {
		t.Errorf("expected description 'A test', got '%s'", created.Description)
	}

	// List items — should include the one we just created
	req = httptest.NewRequest("GET", "/api/items", nil)
	rr = httptest.NewRecorder()

	itemsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("list: expected status 200, got %d", rr.Code)
	}

	var items []Item
	if err := json.Unmarshal(rr.Body.Bytes(), &items); err != nil {
		t.Fatalf("failed to parse items list: %v", err)
	}

	if len(items) == 0 {
		t.Error("expected at least one item in list")
	}
}

func TestItems_GetByID(t *testing.T) {
	// Create an item first
	body := bytes.NewBufferString(`{"name":"Get Test"}`)
	req := httptest.NewRequest("POST", "/api/items", body)
	rr := httptest.NewRecorder()
	itemsHandler(rr, req)

	var created Item
	json.Unmarshal(rr.Body.Bytes(), &created)

	// GET by ID
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/items/%d", created.ID), nil)
	rr = httptest.NewRecorder()
	itemsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var fetched Item
	json.Unmarshal(rr.Body.Bytes(), &fetched)

	if fetched.Name != "Get Test" {
		t.Errorf("expected name 'Get Test', got '%s'", fetched.Name)
	}
}

func TestItems_Update(t *testing.T) {
	// Create an item
	body := bytes.NewBufferString(`{"name":"Before Update"}`)
	req := httptest.NewRequest("POST", "/api/items", body)
	rr := httptest.NewRecorder()
	itemsHandler(rr, req)

	var created Item
	json.Unmarshal(rr.Body.Bytes(), &created)

	// Update it
	body = bytes.NewBufferString(`{"name":"After Update","description":"Updated"}`)
	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/items/%d", created.ID), body)
	rr = httptest.NewRecorder()
	itemsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var updated Item
	json.Unmarshal(rr.Body.Bytes(), &updated)

	if updated.Name != "After Update" {
		t.Errorf("expected name 'After Update', got '%s'", updated.Name)
	}
	if updated.Description != "Updated" {
		t.Errorf("expected description 'Updated', got '%s'", updated.Description)
	}
}

func TestItems_Delete(t *testing.T) {
	// Create an item
	body := bytes.NewBufferString(`{"name":"To Delete"}`)
	req := httptest.NewRequest("POST", "/api/items", body)
	rr := httptest.NewRecorder()
	itemsHandler(rr, req)

	var created Item
	json.Unmarshal(rr.Body.Bytes(), &created)

	// Delete it
	req = httptest.NewRequest("DELETE", fmt.Sprintf("/api/items/%d", created.ID), nil)
	rr = httptest.NewRecorder()
	itemsHandler(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify it's gone
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/items/%d", created.ID), nil)
	rr = httptest.NewRecorder()
	itemsHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404 after delete, got %d", rr.Code)
	}
}

func TestItems_NotFound(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/items/999999", nil)
	rr := httptest.NewRecorder()
	itemsHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestItems_InvalidID(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/items/abc", nil)
	rr := httptest.NewRecorder()
	itemsHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestItems_InvalidJSON(t *testing.T) {
	body := bytes.NewBufferString(`not json`)
	req := httptest.NewRequest("POST", "/api/items", body)
	rr := httptest.NewRecorder()
	itemsHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestItems_MissingName(t *testing.T) {
	body := bytes.NewBufferString(`{"description":"no name"}`)
	req := httptest.NewRequest("POST", "/api/items", body)
	rr := httptest.NewRecorder()
	itemsHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

// =============================================================================
// Display Endpoint Tests
// =============================================================================

func TestDisplay_EmptyByDefault(t *testing.T) {
	resetDisplayData()

	req := httptest.NewRequest("GET", "/api/display", nil)
	rr := httptest.NewRecorder()
	displayHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	if rr.Body.String() != "{}" {
		t.Errorf("expected empty object '{}', got '%s'", rr.Body.String())
	}
}

func TestDisplay_SetAndGet(t *testing.T) {
	resetDisplayData()

	// POST display data
	body := bytes.NewBufferString(`{"terraform":"output","region":"us-east-1"}`)
	req := httptest.NewRequest("POST", "/api/display", body)
	rr := httptest.NewRecorder()
	displayHandler(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("set: expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}

	// GET it back
	req = httptest.NewRequest("GET", "/api/display", nil)
	rr = httptest.NewRecorder()
	displayHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("get: expected status 200, got %d", rr.Code)
	}

	// Parse and verify
	var result map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse display data: %v", err)
	}

	if result["terraform"] != "output" {
		t.Errorf("expected terraform='output', got '%v'", result["terraform"])
	}
	if result["region"] != "us-east-1" {
		t.Errorf("expected region='us-east-1', got '%v'", result["region"])
	}
}

func TestDisplay_InvalidJSON(t *testing.T) {
	body := bytes.NewBufferString(`not json`)
	req := httptest.NewRequest("POST", "/api/display", body)
	rr := httptest.NewRecorder()
	displayHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

// =============================================================================
// System Endpoint Tests
// =============================================================================

func TestSystem_ReturnsExpectedFields(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/system", nil)
	rr := httptest.NewRecorder()
	systemHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse system response: %v", err)
	}

	// Check required fields exist
	for _, field := range []string{"hostname", "ips", "environment", "headers", "client_ip", "user_agent"} {
		if _, ok := result[field]; !ok {
			t.Errorf("expected field '%s' in system response", field)
		}
	}
}

func TestSystem_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/system", nil)
	rr := httptest.NewRecorder()
	systemHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rr.Code)
	}
}
