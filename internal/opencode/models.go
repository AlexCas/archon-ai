package opencode

import (
	"encoding/json"
	"errors"
	"os"
)

// Model represents a single model within a provider.
type Model struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Provider represents a model provider with its model catalog.
type Provider struct {
	ID     string           `json:"id"`
	Name   string           `json:"name"`
	Models map[string]Model `json:"models"`
}

// LoadModelsOrEmpty parses the opencode models cache at cachePath and returns
// a map of providers keyed by provider ID. It tolerates a missing or malformed
// file by returning an empty map — init must never fail due to a missing cache.
func LoadModelsOrEmpty(cachePath string) map[string]Provider {
	if cachePath == "" {
		return map[string]Provider{}
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]Provider{}
		}
		return map[string]Provider{}
	}

	// The cache file is a JSON object keyed by provider ID, each value being
	// a Provider object.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return map[string]Provider{}
	}

	providers := make(map[string]Provider, len(raw))
	for id, provJSON := range raw {
		var p Provider
		if err := json.Unmarshal(provJSON, &p); err != nil {
			continue
		}
		p.ID = id
		providers[id] = p
	}

	return providers
}
