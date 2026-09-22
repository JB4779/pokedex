package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/JB4779/pokedex/internal/pokecache"
)

// Run the real REPL in a child process so exit can call os.Exit safely.
func TestCLIProcess(t *testing.T) {
	mode := os.Getenv("POKEDEX_CLI_TEST_MODE")
	if mode == "" {
		return
	}
	if mode == "main" {
		main()
		os.Exit(0)
	}
	cache := pokecache.NewCache(time.Minute)
	cfg := &config{
		commands: getCommands(), pokedex: make(map[string]Pokemon), cache: cache,
		randomFloat: func() float64 { return 0 },
		httpClient: &http.Client{Transport: mapTransport(func(r *http.Request) (*http.Response, error) {
			name := strings.TrimPrefix(r.URL.Path, "/api/v2/pokemon/")
			name = strings.TrimSuffix(name, "/")
			if name != "pidgey" && name != "caterpie" {
				return nil, fmt.Errorf("unexpected request: %s", r.URL)
			}
			body := fmt.Sprintf(`{"name":%q,"height":3,"weight":18,"base_experience":50,"stats":[{"base_stat":40,"stat":{"name":"hp"}}],"types":[{"type":{"name":"normal"}}]}`, name)
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
		})},
	}
	startRepl(cfg)
	cache.Close()
	os.Exit(0)
}

func TestCLI(t *testing.T) {
	for _, tc := range []struct {
		name, mode, input, wantOut, wantErr string
		contains                            []string
	}{
		{name: "empty Pokedex and exit", mode: "main", input: "pokedex\nexit\npokedex\n", wantOut: "Pokedex > Your Pokedex:\nPokedex > Closing the Pokedex... Goodbye!\n"},
		{name: "normalization and blank input", mode: "main", input: " \t\n  PoKeDeX \t\nEXIT\n", wantOut: "Pokedex > Pokedex > Your Pokedex:\nPokedex > Closing the Pokedex... Goodbye!\n"},
		{name: "EOF", mode: "main", input: "pokedex\n", wantOut: "Pokedex > Your Pokedex:\nPokedex > "},
		{name: "unknown command", mode: "main", input: "missing\nexit\n", wantOut: "Pokedex > Unknown command\nPokedex > Closing the Pokedex... Goodbye!\n"},
		{name: "usage error continues", mode: "main", input: "pokedex extra\npokedex\nexit\n", wantOut: "Pokedex > Pokedex > Your Pokedex:\nPokedex > Closing the Pokedex... Goodbye!\n", wantErr: "Error: usage: pokedex\n"},
		{name: "help registry", mode: "main", input: "help\nexit\n", contains: []string{"Welcome to the Pokedex!", "pokedex: List all caught Pokemon", "Closing the Pokedex... Goodbye!"}},
		{name: "catch list inspect", mode: "fixtures", input: "inspect pidgey\ncatch pidgey\ncatch caterpie\ncatch pidgey\npokedex\ninspect pidgey\nexit\n", wantOut: "Pokedex > you have not caught that pokemon\nPokedex > Throwing a Pokeball at pidgey...\npidgey was caught!\nPokedex > Throwing a Pokeball at caterpie...\ncaterpie was caught!\nPokedex > Throwing a Pokeball at pidgey...\npidgey was caught!\nPokedex > Your Pokedex:\n - caterpie\n - pidgey\nPokedex > Name: pidgey\nHeight: 3\nWeight: 18\nStats:\n  -hp: 40\nTypes:\n  - normal\nPokedex > Closing the Pokedex... Goodbye!\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCLIProcess$")
			cmd.Env = append(os.Environ(), "POKEDEX_CLI_TEST_MODE="+tc.mode)
			cmd.Stdin = strings.NewReader(tc.input)
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("CLI failed: %v\n%s", err, stderr.String())
			}
			if tc.wantOut != "" && stdout.String() != tc.wantOut {
				t.Fatalf("stdout = %q, want %q", stdout.String(), tc.wantOut)
			}
			for _, text := range tc.contains {
				if !strings.Contains(stdout.String(), text) {
					t.Errorf("stdout missing %q: %s", text, stdout.String())
				}
			}
			if stderr.String() != tc.wantErr {
				t.Fatalf("stderr = %q, want %q", stderr.String(), tc.wantErr)
			}
		})
	}
}
