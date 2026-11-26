package main

import (
	"strings"
	"testing"
)

func TestItemInputValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   ItemInput
		wantErr string
	}{
		{
			name:  "valid input",
			input: ItemInput{Name: "Learn Go", Description: "Build a small API"},
		},
		{
			name:  "done flag only",
			input: ItemInput{Name: "Ship it", Done: true},
		},
		{
			name:    "missing name",
			input:   ItemInput{},
			wantErr: "field 'name' is required",
		},
		{
			name:    "whitespace-only name",
			input:   ItemInput{Name: "   "},
			wantErr: "field 'name' is required",
		},
		{
			name:    "name too long",
			input:   ItemInput{Name: strings.Repeat("a", 101)},
			wantErr: "field 'name' must be at most 100 characters",
		},
		{
			name:    "description too long",
			input:   ItemInput{Name: "ok", Description: strings.Repeat("b", 501)},
			wantErr: "field 'description' must be at most 500 characters",
		},
		{
			name:  "max allowed lengths",
			input: ItemInput{Name: strings.Repeat("a", 100), Description: strings.Repeat("b", 500)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := tt.input
			err := in.Validate()

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() = nil, want error %q", tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("Validate() = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestItemInputValidateTrimsFields(t *testing.T) {
	in := ItemInput{
		Name:        "  Learn Go  ",
		Description: "\t Build a small API \n",
	}

	if err := in.Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil", err)
	}
	if in.Name != "Learn Go" {
		t.Errorf("Name = %q, want %q", in.Name, "Learn Go")
	}
	if in.Description != "Build a small API" {
		t.Errorf("Description = %q, want %q", in.Description, "Build a small API")
	}
}
