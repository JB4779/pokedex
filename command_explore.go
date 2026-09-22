package main

import (
	"fmt"
	"net/url"
	"strings"
)

type locationArea struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

func commandExplore(cfg *config, args ...string) error {
	if len(args) != 1 || strings.TrimSpace(args[0]) == "" {
		return fmt.Errorf("usage: explore <area_name>")
	}
	name := args[0]
	fmt.Printf("Exploring %s...\n", name)
	var area locationArea
	if err := fetchJSON(cfg, "https://pokeapi.co/api/v2/location-area/"+url.PathEscape(name)+"/", &area); err != nil {
		return err
	}
	fmt.Println("Found Pokemon:")
	for _, encounter := range area.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}
	return nil
}
