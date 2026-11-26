package main

import (
	"errors"
	"strings"
	"time"
)

// Item is a single record stored in memory.
type Item struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Done        bool      `json:"done"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ItemInput is the payload accepted when creating or updating an item.
type ItemInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

// Validate normalizes the input and checks required fields.
func (in *ItemInput) Validate() error {
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)

	if in.Name == "" {
		return errors.New("field 'name' is required")
	}
	if len(in.Name) > 100 {
		return errors.New("field 'name' must be at most 100 characters")
	}
	if len(in.Description) > 500 {
		return errors.New("field 'description' must be at most 500 characters")
	}
	return nil
}
