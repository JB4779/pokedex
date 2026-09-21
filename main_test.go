package main

import (
	"slices"
	"testing"
)

func TestCleanInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "trimming whitespace",
			input: "  pikachu  ",
			want:  []string{"pikachu"},
		},
		{
			name:  "splitting words",
			input: "catch pikachu now",
			want:  []string{"catch", "pikachu", "now"},
		},
		{
			name:  "lowercasing input",
			input: "CaTcH PIKACHU",
			want:  []string{"catch", "pikachu"},
		},
		{
			name:  "mixed whitespace",
			input: " \tCaTcH\n\r  PIKACHU\t now\n ",
			want:  []string{"catch", "pikachu", "now"},
		},
		{
			name:  "empty input",
			input: "",
			want:  []string{},
		},
		{
			name:  "whitespace only",
			input: " \t\n\r ",
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanInput(tt.input); !slices.Equal(got, tt.want) {
				t.Errorf("cleanInput(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
