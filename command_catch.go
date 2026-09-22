package main

import (
	"fmt"
	"math/rand"
	"net/url"
	"strings"
)

// Pokemon holds the details kept in the user's Pokedex after a catch.
type Pokemon struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	Stats          []struct {
		BaseStat int `json:"base_stat"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
}

func catchChance(baseExperience int) float64 {
	// Missing or zero experience is treated as a low-experience Pokemon.
	return 100 / (100 + float64(max(baseExperience, 1)))
}

func commandCatch(cfg *config, args ...string) error {
	if len(args) != 1 || strings.TrimSpace(args[0]) == "" {
		return fmt.Errorf("usage: catch <pokemon_name>")
	}
	name := args[0]
	var pokemon Pokemon
	if err := fetchJSON(cfg, "https://pokeapi.co/api/v2/pokemon/"+url.PathEscape(name)+"/", &pokemon); err != nil {
		return err
	}
	if pokemon.Name == "" {
		return fmt.Errorf("Pokemon response is missing a name")
	}
	fmt.Printf("Throwing a Pokeball at %s...\n", pokemon.Name)
	roll := cfg.randomFloat
	if roll == nil {
		roll = rand.Float64
	}
	if roll() >= catchChance(pokemon.BaseExperience) {
		fmt.Printf("%s escaped!\n", pokemon.Name)
		return nil
	}
	if cfg.pokedex == nil {
		cfg.pokedex = make(map[string]Pokemon)
	}
	cfg.pokedex[pokemon.Name] = pokemon
	fmt.Printf("%s was caught!\n", pokemon.Name)
	return nil
}
