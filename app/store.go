package main

import (
	"errors"
	"sort"
	"strconv"
	"sync"
	"time"
)

var errNotFound = errors.New("item not found")

// Store keeps items in memory. Safe for concurrent use.
type Store struct {
	mu     sync.RWMutex
	items  map[string]Item
	nextID int
}

func NewStore() *Store {
	return &Store{
		items:  make(map[string]Item),
		nextID: 1,
	}
}

// List returns all items ordered by ID.
func (s *Store) List() []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Item, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		a, _ := strconv.Atoi(items[i].ID)
		b, _ := strconv.Atoi(items[j].ID)
		return a < b
	})
	return items
}

// Get returns a single item by ID.
func (s *Store) Get(id string) (Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[id]
	if !ok {
		return Item{}, errNotFound
	}
	return item, nil
}

// Create inserts a new item and returns it with its generated ID.
func (s *Store) Create(in ItemInput) Item {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	item := Item{
		ID:          strconv.Itoa(s.nextID),
		Name:        in.Name,
		Description: in.Description,
		Done:        in.Done,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.nextID++
	s.items[item.ID] = item
	return item
}

// Update replaces the mutable fields of an existing item.
func (s *Store) Update(id string, in ItemInput) (Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[id]
	if !ok {
		return Item{}, errNotFound
	}

	item.Name = in.Name
	item.Description = in.Description
	item.Done = in.Done
	item.UpdatedAt = time.Now().UTC()
	s.items[id] = item
	return item, nil
}

// Delete removes an item by ID.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[id]; !ok {
		return errNotFound
	}
	delete(s.items, id)
	return nil
}
