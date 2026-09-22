package main

import (
	"fmt"
	"sort"
)

func commandPokedex(cfg *config, args ...string) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: pokedex")
	}
	names := make([]string, 0, len(cfg.pokedex))
	for name := range cfg.pokedex {
		names = append(names, name)
	}
	sort.Strings(names)
	fmt.Println("Your Pokedex:")
	for _, name := range names {
		fmt.Printf(" - %s\n", name)
	}
	return nil
}
