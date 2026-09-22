package main

import (
	"io"
	"os"
	"testing"
)

func TestPokedex(t *testing.T) {
	for _, tc := range []struct {
		name    string
		pokedex map[string]Pokemon
		args    []string
		want    string
		wantErr bool
	}{
		{"empty", map[string]Pokemon{}, nil, "Your Pokedex:\n", false},
		{"nil", nil, nil, "Your Pokedex:\n", false},
		{"caught Pokemon", map[string]Pokemon{"pidgey": {Name: "pidgey"}, "caterpie": {Name: "caterpie"}}, nil, "Your Pokedex:\n - caterpie\n - pidgey\n", false},
		{"extra argument", nil, []string{"pidgey"}, "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output, err := os.CreateTemp(t.TempDir(), "stdout")
			if err != nil {
				t.Fatal(err)
			}
			defer output.Close()
			original := os.Stdout
			os.Stdout = output
			defer func() { os.Stdout = original }()
			// Listing caught Pokemon requires neither an HTTP client nor a cache.
			err = getCommands()["pokedex"].callback(&config{pokedex: tc.pokedex}, tc.args...)
			os.Stdout = original
			if (err != nil) != tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr && err.Error() != "usage: pokedex" {
				t.Fatalf("unexpected usage: %v", err)
			}
			if _, err := output.Seek(0, 0); err != nil {
				t.Fatal(err)
			}
			got, err := io.ReadAll(output)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.want {
				t.Fatalf("output = %q, want %q", got, tc.want)
			}
		})
	}
}
