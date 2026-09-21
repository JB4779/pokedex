package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

func displayLocationAreas(cfg *config, pageURL string) error {
	// PokeAPI includes offset=0 in back links but omits it in the initial URL.
	parsed, err := url.Parse(pageURL)
	if err != nil {
		return fmt.Errorf("parse location areas URL: %w", err)
	}
	query := parsed.Query()
	if query.Get("offset") == "0" {
		query.Del("offset")
	}
	parsed.RawQuery = query.Encode()
	key := parsed.String()
	var data []byte
	var cached bool
	if cfg.cache != nil {
		data, cached = cfg.cache.Get(key)
	}
	if !cached {
		resp, err := cfg.httpClient.Get(pageURL)
		if err != nil {
			return fmt.Errorf("fetch location areas: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("fetch location areas: %s", resp.Status)
		}
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read location areas: %w", err)
		}
	}
	var page locationAreaPage
	if err := json.Unmarshal(data, &page); err != nil {
		return fmt.Errorf("decode location areas: %w", err)
	}
	if !cached && cfg.cache != nil {
		cfg.cache.Add(key, data)
	}
	cfg.Next, cfg.Previous = page.Next, page.Previous
	for _, area := range page.Results {
		fmt.Println(area.Name)
	}
	return nil
}
