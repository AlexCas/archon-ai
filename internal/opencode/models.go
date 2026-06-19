package opencode

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Model is one model entry within a provider. Only the fields the foundation
// needs are captured; the cache carries many more (family/cost/limit/…) which
// encoding/json silently ignores.
type Model struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ToolCall  bool   `json:"tool_call"`
	Reasoning bool   `json:"reasoning"`
}

// Provider is one provider in the opencode cache. Models are keyed by the
// cache's model key: BARE under the "opencode" provider (e.g. "deepseek-v4-pro")
// and ALREADY-SLASHED under other providers (e.g. "xai/grok-4").
type Provider struct {
	ID     string           `json:"id"`
	Name   string           `json:"name"`
	Models map[string]Model `json:"models"`
}

// DefaultCachePath returns ~/.cache/opencode/models.json for the current user.
// Unlike gentle-ai's version it returns the error (archon convention) rather
// than swallowing it to "".
func DefaultCachePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cache", "opencode", "models.json"), nil
}

// LoadModels parses the provider-keyed cache into map[providerID]Provider.
// A malformed/partial provider entry is skipped (not fatal); a malformed
// top-level JSON document is an error. The map key is forced onto p.ID so the
// provider id is authoritative even if the inner "id" is missing/wrong.
func LoadModels(path string) (map[string]Provider, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read models cache %q: %w", path, err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse models cache: %w", err)
	}
	providers := make(map[string]Provider, len(raw))
	for id, pjson := range raw {
		var p Provider
		if err := json.Unmarshal(pjson, &p); err != nil {
			continue // skip malformed entry (C2)
		}
		p.ID = id
		providers[id] = p
	}
	return providers, nil
}

// LoadModelsOrEmpty returns an empty map and nil error when the cache file is
// absent; any other read/parse error from LoadModels propagates.
func LoadModelsOrEmpty(path string) (map[string]Provider, error) {
	providers, err := LoadModels(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]Provider{}, nil
		}
		return nil, err
	}
	return providers, nil
}
