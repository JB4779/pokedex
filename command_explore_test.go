package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/JB4779/pokedex/internal/pokecache"
)

func TestExplore(t *testing.T) {
	for _, tc := range []struct {
		name, body, output string
		status             int
		fail               bool
	}{
		{"encounters", `{"pokemon_encounters":[{"pokemon":{"name":"tentacool"}},{"pokemon":{"name":"tentacruel"}}]}`, "Found Pokemon:\n - tentacool\n - tentacruel\n", 200, false},
		{"empty", `{"pokemon_encounters":[]}`, "Found Pokemon:\n", 200, false},
		{"not found", `{}`, "", 404, true},
		{"invalid JSON", `{`, "", 200, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cache := pokecache.NewCache(time.Minute)
			defer cache.Close()
			requests := 0
			next, previous := "next", "previous"
			cfg := &config{Next: &next, Previous: &previous, cache: cache, httpClient: &http.Client{Transport: mapTransport(func(r *http.Request) (*http.Response, error) {
				requests++
				if r.URL.String() != "https://pokeapi.co/api/v2/location-area/pastoria-city-area/" {
					t.Fatalf("wrong URL: %s", r.URL)
				}
				return &http.Response{StatusCode: tc.status, Status: fmt.Sprint(tc.status), Body: io.NopCloser(strings.NewReader(tc.body))}, nil
			})}}
			output, err := os.CreateTemp(t.TempDir(), "stdout")
			if err != nil {
				t.Fatal(err)
			}
			defer output.Close()
			original := os.Stdout
			os.Stdout = output
			defer func() { os.Stdout = original }()
			for i := 0; i < 2; i++ {
				err := commandExplore(cfg, "pastoria-city-area")
				if (err != nil) != tc.fail {
					t.Fatalf("unexpected error: %v", err)
				}
			}
			os.Stdout = original
			if _, err := output.Seek(0, 0); err != nil {
				t.Fatal(err)
			}
			data, err := io.ReadAll(output)
			if err != nil {
				t.Fatal(err)
			}
			want := strings.Repeat("Exploring pastoria-city-area...\n"+tc.output, 2)
			if string(data) != want {
				t.Fatalf("output = %q, want %q", data, want)
			}
			expectedRequests := 1
			if tc.fail {
				expectedRequests = 2
			}
			if requests != expectedRequests {
				t.Fatalf("requests = %d, want %d", requests, expectedRequests)
			}
			if cfg.Next != &next || cfg.Previous != &previous {
				t.Fatal("explore changed map pagination")
			}
		})
	}
}

func TestExploreArguments(t *testing.T) {
	for _, args := range [][]string{nil, {""}, {"a", "b"}} {
		if err := commandExplore(&config{}, args...); err == nil || !strings.Contains(err.Error(), "usage: explore <area_name>") {
			t.Fatalf("args %v: got %v", args, err)
		}
	}
}
