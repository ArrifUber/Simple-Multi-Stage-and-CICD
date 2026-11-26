package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer() http.Handler {
	return NewAPI(NewStore()).Routes()
}

// do sends a JSON-encoded body (nil for no body) and returns the recorder.
func do(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	return doRaw(t, h, method, path, buf.String())
}

// doRaw sends a raw request body.
func doRaw(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// decode decodes a JSON response body into T.
func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()

	var v T
	if err := json.NewDecoder(rec.Body).Decode(&v); err != nil {
		t.Fatalf("decode response %q: %v", rec.Body.String(), err)
	}
	return v
}

func assertJSONContentType(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
}

func TestCRUD(t *testing.T) {
	h := newTestServer()

	// Create
	rec := do(t, h, http.MethodPost, "/items", ItemInput{Name: "Learn Go"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: got %d, want %d", rec.Code, http.StatusCreated)
	}
	assertJSONContentType(t, rec)
	created := decode[Item](t, rec)
	if created.ID != "1" || created.Name != "Learn Go" {
		t.Fatalf("unexpected created item: %+v", created)
	}
	if loc := rec.Header().Get("Location"); loc != "/items/1" {
		t.Fatalf("Location = %q, want %q", loc, "/items/1")
	}

	// List
	rec = do(t, h, http.MethodGet, "/items", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list: got %d, want %d", rec.Code, http.StatusOK)
	}
	assertJSONContentType(t, rec)
	items := decode[[]Item](t, rec)
	if len(items) != 1 {
		t.Fatalf("list: got %d items, want 1", len(items))
	}

	// Get
	rec = do(t, h, http.MethodGet, "/items/1", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get: got %d, want %d", rec.Code, http.StatusOK)
	}
	if got := decode[Item](t, rec); got.ID != created.ID {
		t.Fatalf("get: ID = %q, want %q", got.ID, created.ID)
	}

	// Update
	rec = do(t, h, http.MethodPut, "/items/1", ItemInput{Name: "Learn Go", Done: true})
	if rec.Code != http.StatusOK {
		t.Fatalf("update: got %d, want %d", rec.Code, http.StatusOK)
	}
	updated := decode[Item](t, rec)
	if !updated.Done {
		t.Fatal("update: done should be true")
	}
	if updated.UpdatedAt.Before(created.UpdatedAt) {
		t.Fatalf("update: UpdatedAt went backwards")
	}

	// Delete
	rec = do(t, h, http.MethodDelete, "/items/1", nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete: got %d, want %d", rec.Code, http.StatusNoContent)
	}
	if body := rec.Body.String(); body != "" {
		t.Errorf("delete: body = %q, want empty", body)
	}

	// Get after delete
	rec = do(t, h, http.MethodGet, "/items/1", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get deleted: got %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestHealth(t *testing.T) {
	h := newTestServer()

	rec := do(t, h, http.MethodGet, "/health", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	assertJSONContentType(t, rec)

	body := decode[map[string]string](t, rec)
	if body["status"] != "ok" {
		t.Errorf(`body["status"] = %q, want "ok"`, body["status"])
	}
}

func TestIndex(t *testing.T) {
	h := newTestServer()

	rec := do(t, h, http.MethodGet, "/", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	assertJSONContentType(t, rec)

	body := decode[struct {
		Name   string   `json:"name"`
		Routes []string `json:"routes"`
	}](t, rec)
	if body.Name != "simple-api" {
		t.Errorf("name = %q, want %q", body.Name, "simple-api")
	}
	if len(body.Routes) == 0 {
		t.Error("routes is empty, want documented endpoints")
	}
}

func TestNotFoundAndMethodNotAllowed(t *testing.T) {
	h := newTestServer()
	do(t, h, http.MethodPost, "/items", ItemInput{Name: "existing"})

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		want   int
	}{
		{"unknown route", http.MethodGet, "/nope", "", http.StatusNotFound},
		{"nested path", http.MethodGet, "/items/1/extra", "", http.StatusNotFound},
		{"get missing item", http.MethodGet, "/items/999", "", http.StatusNotFound},
		{"update missing item", http.MethodPut, "/items/999", `{"name":"x"}`, http.StatusNotFound},
		{"delete missing item", http.MethodDelete, "/items/999", "", http.StatusNotFound},
		{"patch not supported", http.MethodPatch, "/items/1", "", http.StatusMethodNotAllowed},
		{"post to item not supported", http.MethodPost, "/items/1", `{"name":"x"}`, http.StatusMethodNotAllowed},
		{"put on collection not supported", http.MethodPut, "/items", `{"name":"x"}`, http.StatusMethodNotAllowed},
		{"delete on collection not supported", http.MethodDelete, "/items", "", http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRaw(t, h, tt.method, tt.path, tt.body)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestInvalidCreateRequests(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"malformed json", "{bad json"},
		{"unknown field", `{"name":"x","color":"red"}`},
		{"multiple objects", `{"name":"a"}{"name":"b"}`},
		{"missing name", `{"description":"x"}`},
		{"whitespace-only name", `{"name":"   "}`},
		{"name too long", `{"name":"` + strings.Repeat("a", 101) + `"}`},
		{"description too long", `{"name":"ok","description":"` + strings.Repeat("b", 501) + `"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestServer()
			rec := doRaw(t, h, http.MethodPost, "/items", tt.body)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
			assertJSONContentType(t, rec)

			body := decode[map[string]string](t, rec)
			if body["error"] == "" {
				t.Error(`response has no "error" message`)
			}

			// Nothing should have been stored.
			if items := decode[[]Item](t, do(t, h, http.MethodGet, "/items", nil)); len(items) != 0 {
				t.Errorf("store has %d items after invalid request, want 0", len(items))
			}
		})
	}
}

func TestBodyTooLarge(t *testing.T) {
	h := newTestServer()

	body := `{"name":"` + strings.Repeat("a", 1<<20) + `"}`
	rec := doRaw(t, h, http.MethodPost, "/items", body)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if msg := decode[map[string]string](t, rec)["error"]; !strings.Contains(msg, "too large") {
		t.Errorf("error = %q, want mention of oversized body", msg)
	}
}

func TestUpdateValidationKeepsOldValues(t *testing.T) {
	h := newTestServer()
	created := decode[Item](t, do(t, h, http.MethodPost, "/items", ItemInput{Name: "original"}))

	rec := do(t, h, http.MethodPut, "/items/"+created.ID, ItemInput{Name: "   "})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	got := decode[Item](t, do(t, h, http.MethodGet, "/items/"+created.ID, nil))
	if got.Name != "original" {
		t.Errorf("Name = %q, want %q (invalid update must not change data)", got.Name, "original")
	}
}
