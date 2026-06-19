package opencode

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestDefaultCachePath asserts the path ends with .cache/opencode/models.json
// relative to the injected HOME. (C1)
func TestDefaultCachePath(t *testing.T) {
	t.Setenv("HOME", "/fakehome")
	got, err := DefaultCachePath()
	if err != nil {
		t.Fatalf("DefaultCachePath() error = %v", err)
	}
	want := filepath.Join("/fakehome", ".cache", "opencode", "models.json")
	if got != want {
		t.Errorf("DefaultCachePath() = %q, want %q", got, want)
	}
	if !strings.HasSuffix(got, filepath.Join(".cache", "opencode", "models.json")) {
		t.Errorf("DefaultCachePath() = %q does not end with .cache/opencode/models.json", got)
	}
}

// TestLoadModels_WellFormed loads testdata/models.json and asserts the happy
// path: opencode provider present, models keyed correctly, p.ID == map key for
// all providers, slashed key preserved for second provider. (C2 happy)
func TestLoadModels_WellFormed(t *testing.T) {
	providers, err := LoadModels("testdata/models.json")
	if err != nil {
		t.Fatalf("LoadModels() error = %v", err)
	}

	// opencode provider must be present
	opencode, ok := providers["opencode"]
	if !ok {
		t.Fatalf("providers[\"opencode\"] missing")
	}
	if opencode.ID != "opencode" {
		t.Errorf("opencode.ID = %q, want %q", opencode.ID, "opencode")
	}

	// bare-key models under opencode
	if _, ok := opencode.Models["deepseek-v4-pro"]; !ok {
		t.Error("opencode.Models[\"deepseek-v4-pro\"] missing")
	}
	if _, ok := opencode.Models["glm-5"]; !ok {
		t.Error("opencode.Models[\"glm-5\"] missing")
	}

	// tool_call=true preserved
	if !opencode.Models["deepseek-v4-pro"].ToolCall {
		t.Error("opencode.Models[\"deepseek-v4-pro\"].ToolCall = false, want true")
	}

	// p.ID == map key for all providers
	for id, p := range providers {
		if p.ID != id {
			t.Errorf("providers[%q].ID = %q, want map key", id, p.ID)
		}
	}

	// requesty provider: slashed key preserved
	requesty, ok := providers["requesty"]
	if !ok {
		t.Fatalf("providers[\"requesty\"] missing")
	}
	if _, ok := requesty.Models["xai/grok-4"]; !ok {
		t.Error("requesty.Models[\"xai/grok-4\"] missing (slashed key not preserved)")
	}
}

// TestLoadModels_MalformedEntrySkipped loads testdata/malformed.json: good
// provider returned, broken entry absent, nil error. (C2 malformed-skipped)
func TestLoadModels_MalformedEntrySkipped(t *testing.T) {
	providers, err := LoadModels("testdata/malformed.json")
	if err != nil {
		t.Fatalf("LoadModels() error = %v, want nil", err)
	}
	if _, ok := providers["opencode"]; !ok {
		t.Error("providers[\"opencode\"] missing — good entry should be returned")
	}
	if _, ok := providers["broken"]; ok {
		t.Error("providers[\"broken\"] present — malformed entry should be skipped")
	}
}

// TestLoadModelsOrEmpty_Absent calls LoadModelsOrEmpty with a non-existent path
// and asserts empty map and nil error. (C3 absent)
func TestLoadModelsOrEmpty_Absent(t *testing.T) {
	providers, err := LoadModelsOrEmpty("/nonexistent/does/not/exist/models.json")
	if err != nil {
		t.Fatalf("LoadModelsOrEmpty() error = %v, want nil", err)
	}
	if len(providers) != 0 {
		t.Errorf("LoadModelsOrEmpty() returned %d providers, want 0", len(providers))
	}
}

// TestLoadModelsOrEmpty_ParseError calls LoadModelsOrEmpty with testdata/invalid.json
// and asserts a non-nil error. (C3 parse-error)
func TestLoadModelsOrEmpty_ParseError(t *testing.T) {
	_, err := LoadModelsOrEmpty("testdata/invalid.json")
	if err == nil {
		t.Error("LoadModelsOrEmpty(invalid.json) error = nil, want non-nil")
	}
}
