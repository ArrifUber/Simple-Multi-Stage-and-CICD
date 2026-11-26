package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer() http.Handler {
	return NewAPI(NewStore()).Routes()
}

func do(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestCRUD(t *testing.T) {
	h := newTestServer()

	// Create
	rec := do(t, h, http.MethodPost, "/items", ItemInput{Name: "Learn Go"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: got %d, want %d", rec.Code, http.StatusCreated)
	}
	var created Item
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	if created.ID != "1" || created.Name != "Learn Go" {
		t.Fatalf("unexpected created item: %+v", created)
	}
	if loc := rec.Header().Get("Location"); loc != "/items/1" {
		t.Fatalf("location: got %q", loc)
	}

	// List
	rec = do(t, h, http.MethodGet, "/items", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list: got %d", rec.Code)
	}
	var items []Item
	if err := json.NewDecoder(rec.Body).Decode(&items); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("list: got %d items, want 1", len(items))
	}

	// Update
	rec = do(t, h, http.MethodPut, "/items/1", ItemInput{Name: "Learn Go", Done: true})
	if rec.Code != http.StatusOK {
		t.Fatalf("update: got %d", rec.Code)
	}
	var updated Item
	if err := json.NewDecoder(rec.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated: %v", err)
	}
	if !updated.Done {
		t.Fatal("update: done should be true")
	}

	// Delete
	rec = do(t, h, http.MethodDelete, "/items/1", nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete: got %d", rec.Code)
	}

	// Get after delete
	rec = do(t, h, http.MethodGet, "/items/1", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get deleted: got %d, want 404", rec.Code)
	}
}

func TestValidation(t *testing.T) {
	h := newTestServer()

	rec := do(t, h, http.MethodPost, "/items", ItemInput{})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("empty name: got %d, want 400", rec.Code)
	}

	req := httptest.NewRequest(http.MethodPost, "/items", bytes.NewBufferString("{bad json"))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad json: got %d, want 400", rec.Code)
	}
}
