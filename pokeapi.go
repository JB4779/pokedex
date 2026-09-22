package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// fetchJSON caches only responses that successfully decode into target.
func fetchJSON(cfg *config, pageURL string, target any) error {
	// PokeAPI includes offset=0 in back links but omits it in the initial URL.
	parsed, err := url.Parse(pageURL)
	if err != nil {
		return fmt.Errorf("parse PokeAPI response URL: %w", err)
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
			return fmt.Errorf("fetch PokeAPI response: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("fetch PokeAPI response: %s", resp.Status)
		}
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read PokeAPI response: %w", err)
		}
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode PokeAPI response: %w", err)
	}
	if !cached && cfg.cache != nil {
		cfg.cache.Add(key, data)
	}
	return nil
}
