package main

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/JB4779/pokedex/internal/pokecache"
)

func TestCatchOutcomesAndCache(t *testing.T) {
	cache := pokecache.NewCache(time.Minute)
	defer cache.Close()
	requests := 0
	cfg := &config{cache: cache, httpClient: &http.Client{Transport: mapTransport(func(r *http.Request) (*http.Response, error) {
		requests++
		if r.URL.String() != "https://pokeapi.co/api/v2/pokemon/pikachu/" {
			t.Fatalf("unexpected URL %s", r.URL)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"id":25,"name":"pikachu","base_experience":112,"height":4,"weight":60}`))}, nil
	})}}
	output, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	original := os.Stdout
	os.Stdout = output
	defer func() { os.Stdout = original }()
	for _, tc := range []struct {
		roll   float64
		caught bool
	}{{0.99, false}, {0, true}, {0.99, true}, {0, true}} {
		cfg.randomFloat = func() float64 { return tc.roll }
		if err := commandCatch(cfg, "pikachu"); err != nil {
			t.Fatal(err)
		}
		pokemon, ok := cfg.pokedex["pikachu"]
		if ok != tc.caught {
			t.Fatalf("caught = %v, want %v", ok, tc.caught)
		}
		if ok && (pokemon.ID != 25 || pokemon.BaseExperience != 112 || pokemon.Weight != 60) {
			t.Fatalf("lost Pokemon details: %+v", pokemon)
		}
	}
	if requests != 1 || len(cfg.pokedex) != 1 {
		t.Fatalf("requests=%d, pokedex size=%d", requests, len(cfg.pokedex))
	}
	os.Stdout = original
	output.Seek(0, 0)
	data, err := io.ReadAll(output)
	if err != nil {
		t.Fatal(err)
	}
	want := "Throwing a Pokeball at pikachu...\npikachu escaped!\nThrowing a Pokeball at pikachu...\npikachu was caught!\n"
	if string(data) != want+want {
		t.Fatalf("unexpected output: %s", data)
	}
}

func TestCatchChance(t *testing.T) {
	previous := 1.0
	for _, experience := range []int{1, 50, 112, 300, 600} {
		chance := catchChance(experience)
		if chance <= 0 || chance >= previous {
			t.Fatalf("invalid chance %f for %d", chance, experience)
		}
		previous = chance
	}
	for _, experience := range []int{0, -1} {
		if chance := catchChance(experience); chance <= 0 || chance >= 1 {
			t.Fatalf("invalid chance: %f", chance)
		}
	}
}

func TestCatchInvalidArguments(t *testing.T) {
	for _, args := range [][]string{nil, {""}, {" "}, {"pikachu", "extra"}} {
		if err := commandCatch(&config{}, args...); err == nil {
			t.Fatalf("expected error for %v", args)
		}
	}
}

func TestCatchAPIFailures(t *testing.T) {
	for _, tc := range []struct {
		status int
		body   string
	}{{404, `{}`}, {200, `{`}, {200, `{}`}} {
		cfg := &config{httpClient: &http.Client{Transport: mapTransport(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
		})}, randomFloat: func() float64 { t.Fatal("rolled after API failure"); return 0 }}
		if err := commandCatch(cfg, "pikachu"); err == nil {
			t.Fatal("expected error")
		}
		if len(cfg.pokedex) != 0 {
			t.Fatal("caught Pokemon after API failure")
		}
	}
}
