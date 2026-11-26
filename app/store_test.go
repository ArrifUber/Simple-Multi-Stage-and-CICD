package main

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestStoreCreate(t *testing.T) {
	s := NewStore()

	first := s.Create(ItemInput{Name: "Learn Go", Description: "Build an API"})
	second := s.Create(ItemInput{Name: "Ship it", Done: true})

	if first.ID != "1" {
		t.Errorf("first.ID = %q, want %q", first.ID, "1")
	}
	if second.ID != "2" {
		t.Errorf("second.ID = %q, want %q", second.ID, "2")
	}
	if first.Name != "Learn Go" || first.Description != "Build an API" || first.Done {
		t.Errorf("first item not stored as given: %+v", first)
	}
	if !second.Done {
		t.Errorf("second.Done = false, want true")
	}
	if first.CreatedAt.IsZero() || first.UpdatedAt.IsZero() {
		t.Errorf("timestamps not set: created=%v updated=%v", first.CreatedAt, first.UpdatedAt)
	}
	if !first.CreatedAt.Equal(first.UpdatedAt) {
		t.Errorf("new item timestamps differ: created=%v updated=%v", first.CreatedAt, first.UpdatedAt)
	}
	if first.CreatedAt.Location() != time.UTC {
		t.Errorf("CreatedAt location = %v, want UTC", first.CreatedAt.Location())
	}
}

func TestStoreGet(t *testing.T) {
	s := NewStore()
	created := s.Create(ItemInput{Name: "Learn Go"})

	got, err := s.Get(created.ID)
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got.ID != created.ID || got.Name != created.Name {
		t.Errorf("Get() = %+v, want %+v", got, created)
	}

	if _, err := s.Get("does-not-exist"); !errors.Is(err, errNotFound) {
		t.Errorf("Get(missing) error = %v, want errNotFound", err)
	}
}

func TestStoreListIsEmptySlice(t *testing.T) {
	items := NewStore().List()

	if items == nil {
		t.Fatal("List() = nil, want empty non-nil slice (encodes as [])")
	}
	if len(items) != 0 {
		t.Fatalf("List() returned %d items, want 0", len(items))
	}
}

func TestStoreListOrdersByNumericID(t *testing.T) {
	s := NewStore()
	for i := 0; i < 12; i++ {
		s.Create(ItemInput{Name: fmt.Sprintf("item-%d", i)})
	}

	items := s.List()
	if len(items) != 12 {
		t.Fatalf("List() returned %d items, want 12", len(items))
	}
	for i, item := range items {
		want := fmt.Sprintf("%d", i+1)
		if item.ID != want {
			t.Fatalf("items[%d].ID = %q, want %q (list must be numeric, not lexicographic)", i, item.ID, want)
		}
	}
}

func TestStoreUpdate(t *testing.T) {
	s := NewStore()
	created := s.Create(ItemInput{Name: "Learn Go"})

	updated, err := s.Update(created.ID, ItemInput{Name: "Learn Go", Description: "Done", Done: true})
	if err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
	if updated.Name != "Learn Go" || updated.Description != "Done" || !updated.Done {
		t.Errorf("Update() = %+v, fields not applied", updated)
	}
	if !updated.CreatedAt.Equal(created.CreatedAt) {
		t.Errorf("CreatedAt changed: got %v, want %v", updated.CreatedAt, created.CreatedAt)
	}
	if updated.UpdatedAt.Before(created.UpdatedAt) {
		t.Errorf("UpdatedAt went backwards: got %v, before %v", updated.UpdatedAt, created.UpdatedAt)
	}

	// The update must persist.
	stored, err := s.Get(created.ID)
	if err != nil {
		t.Fatalf("Get() after update error = %v", err)
	}
	if !stored.Done || stored.Description != "Done" {
		t.Errorf("stored item = %+v, want updated values", stored)
	}

	if _, err := s.Update("missing", ItemInput{Name: "x"}); !errors.Is(err, errNotFound) {
		t.Errorf("Update(missing) error = %v, want errNotFound", err)
	}
}

func TestStoreDelete(t *testing.T) {
	s := NewStore()
	created := s.Create(ItemInput{Name: "Learn Go"})

	if err := s.Delete(created.ID); err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}
	if _, err := s.Get(created.ID); !errors.Is(err, errNotFound) {
		t.Errorf("Get() after delete error = %v, want errNotFound", err)
	}
	if len(s.List()) != 0 {
		t.Errorf("List() after delete returned %d items, want 0", len(s.List()))
	}

	if err := s.Delete(created.ID); !errors.Is(err, errNotFound) {
		t.Errorf("Delete() twice error = %v, want errNotFound", err)
	}
}

func TestStoreConcurrentAccess(t *testing.T) {
	s := NewStore()

	const workers = 100
	ids := make(chan string, workers)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			item := s.Create(ItemInput{Name: fmt.Sprintf("item-%d", i)})
			ids <- item.ID

			if _, err := s.Get(item.ID); err != nil {
				t.Errorf("Get(%q) error = %v", item.ID, err)
			}
			_ = s.List()
			if _, err := s.Update(item.ID, ItemInput{Name: "updated", Done: true}); err != nil {
				t.Errorf("Update(%q) error = %v", item.ID, err)
			}
		}(i)
	}
	wg.Wait()
	close(ids)

	seen := make(map[string]bool, workers)
	for id := range ids {
		if seen[id] {
			t.Errorf("duplicate ID generated: %q", id)
		}
		seen[id] = true
	}
	if len(seen) != workers {
		t.Errorf("got %d unique IDs, want %d", len(seen), workers)
	}
	if got := len(s.List()); got != workers {
		t.Errorf("List() returned %d items, want %d", got, workers)
	}
}
