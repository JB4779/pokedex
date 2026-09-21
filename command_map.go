package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type locationAreaPage struct {
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
	} `json:"results"`
}

func commandMap(cfg *config) error {
	if cfg.Next == nil {
		fmt.Println("you're on the last page")
		return nil
	}
	return displayLocationAreas(cfg, *cfg.Next)
}

func commandMapb(cfg *config) error {
	if cfg.Previous == nil {
		fmt.Println("you're on the first page")
		return nil
	}
	return displayLocationAreas(cfg, *cfg.Previous)
}

func displayLocationAreas(cfg *config, url string) error {
	resp, err := cfg.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("fetch location areas: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch location areas: %s", resp.Status)
	}
	var page locationAreaPage
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return fmt.Errorf("decode location areas: %w", err)
	}
	cfg.Next, cfg.Previous = page.Next, page.Previous
	for _, area := range page.Results {
		fmt.Println(area.Name)
	}
	return nil
}
