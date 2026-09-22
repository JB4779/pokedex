package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/JB4779/pokedex/internal/pokecache"
)

type mapTransport func(*http.Request) (*http.Response, error)

func (f mapTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestMapPagination(t *testing.T) {
	first, second := "https://example.test/?limit=20", "https://example.test/?limit=20&offset=20"
	back := first + "&offset=0"
	var requested []string
	cache := pokecache.NewCache(time.Minute)
	defer cache.Close()
	cfg := &config{cache: cache, Next: &first, httpClient: &http.Client{Transport: mapTransport(func(r *http.Request) (*http.Response, error) {
		requested = append(requested, r.URL.String())
		body := fmt.Sprintf(`{"next":%q,"previous":null,"results":[{"name":"canalave-city-area"}]}`, second)
		if r.URL.String() == second {
			body = fmt.Sprintf(`{"next":null,"previous":%q,"results":[{"name":"great-marsh-area-1"}]}`, first+"&offset=0")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}}
	for _, step := range []struct {
		name           string
		callback       func(*config, ...string) error
		next, previous *string
		requests       int
	}{
		{"back before first", commandMapb, &first, nil, 0},
		{"first", commandMap, &second, nil, 1},
		{"back on first", commandMapb, &second, nil, 1},
		{"second", commandMap, nil, &back, 2},
		{"last boundary", commandMap, nil, &back, 2},
		{"back", commandMapb, &second, nil, 2},
		{"forward cached", commandMap, nil, &back, 2},
	} {
		t.Run(step.name, func(t *testing.T) {
			if err := step.callback(cfg); err != nil {
				t.Fatal(err)
			}
			equal := func(a, b *string) bool { return a == nil && b == nil || a != nil && b != nil && *a == *b }
			if !equal(cfg.Next, step.next) || !equal(cfg.Previous, step.previous) {
				t.Fatal("incorrect pagination state")
			}
			if len(requested) != step.requests {
				t.Fatalf("got %d requests, want %d", len(requested), step.requests)
			}
		})
	}
	if strings.Join(requested, "|") != strings.Join([]string{first, second}, "|") {
		t.Fatalf("wrong request order: %v", requested)
	}
}

func TestMapErrorsPreservePagination(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		err    error
	}{
		{"HTTP error", 500, `{}`, nil},
		{"invalid JSON", 200, `{`, nil},
		{"network error", 0, "", fmt.Errorf("connection failed")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cache := pokecache.NewCache(time.Minute)
			defer cache.Close()
			next, previous := "https://example.test/next", "https://example.test/previous"
			cfg := &config{cache: cache, Next: &next, Previous: &previous, httpClient: &http.Client{Transport: mapTransport(func(r *http.Request) (*http.Response, error) {
				if tc.err != nil {
					return nil, tc.err
				}
				return &http.Response{StatusCode: tc.status, Status: fmt.Sprint(tc.status), Body: io.NopCloser(strings.NewReader(tc.body))}, nil
			})}}
			for _, callback := range []func(*config, ...string) error{commandMap, commandMapb, commandMap, commandMapb} {
				if err := callback(cfg); err == nil {
					t.Fatal("expected error")
				}
				if cfg.Next != &next || cfg.Previous != &previous {
					t.Fatal("pagination changed after failure")
				}
			}
		})
	}
}
