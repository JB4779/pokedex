package main

import (
	"fmt"
)

type locationAreaPage struct {
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
	} `json:"results"`
}

func commandMap(cfg *config, _ ...string) error {
	if cfg.Next == nil {
		fmt.Println("you're on the last page")
		return nil
	}
	return displayLocationAreas(cfg, *cfg.Next)
}

func commandMapb(cfg *config, _ ...string) error {
	if cfg.Previous == nil {
		fmt.Println("you're on the first page")
		return nil
	}
	return displayLocationAreas(cfg, *cfg.Previous)
}

func displayLocationAreas(cfg *config, pageURL string) error {
	var page locationAreaPage
	if err := fetchJSON(cfg, pageURL, &page); err != nil {
		return err
	}
	cfg.Next, cfg.Previous = page.Next, page.Previous
	for _, area := range page.Results {
		fmt.Println(area.Name)
	}
	return nil
}
