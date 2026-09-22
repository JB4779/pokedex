package main

import (
	"encoding/json"
	"io"
	"os"
	"testing"
)

func TestInspect(t *testing.T) {
	var pidgey Pokemon
	err := json.Unmarshal([]byte(`{"name":"pidgey","height":3,"weight":18,"stats":[{"base_stat":40,"stat":{"name":"hp"}},{"base_stat":45,"stat":{"name":"attack"}},{"base_stat":40,"stat":{"name":"defense"}},{"base_stat":35,"stat":{"name":"special-attack"}},{"base_stat":35,"stat":{"name":"special-defense"}},{"base_stat":56,"stat":{"name":"speed"}}],"types":[{"type":{"name":"normal"}},{"type":{"name":"flying"}}]}`), &pidgey)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		pokedex map[string]Pokemon
		args    []string
		want    string
		wantErr bool
	}{
		{"caught", map[string]Pokemon{"pidgey": pidgey}, []string{"pidgey"}, "Name: pidgey\nHeight: 3\nWeight: 18\nStats:\n  -hp: 40\n  -attack: 45\n  -defense: 40\n  -special-attack: 35\n  -special-defense: 35\n  -speed: 56\nTypes:\n  - normal\n  - flying\n", false},
		{"uncaught", map[string]Pokemon{"pidgey": pidgey}, []string{"pikachu"}, "you have not caught that pokemon\n", false},
		{"empty pokedex", nil, []string{"pidgey"}, "you have not caught that pokemon\n", false},
		{"missing argument", nil, nil, "", true},
		{"blank argument", nil, []string{" "}, "", true},
		{"extra argument", nil, []string{"pidgey", "pikachu"}, "", true},
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
			// No HTTP client or cache: inspection must use only caught Pokemon.
			cfg := &config{pokedex: tc.pokedex}
			err = getCommands()["inspect"].callback(cfg, tc.args...)
			os.Stdout = original
			if (err != nil) != tc.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr && err.Error() != "usage: inspect <pokemon_name>" {
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
