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

// TestHasToolCallModel covers the tool_call detection helper.
func TestHasToolCallModel(t *testing.T) {
	withTC := Provider{ID: "p", Models: map[string]Model{
		"a": {Name: "a", ToolCall: false},
		"b": {Name: "b", ToolCall: true},
	}}
	if !hasToolCallModel(withTC) {
		t.Error("hasToolCallModel(withTC) = false, want true")
	}
	noTC := Provider{ID: "p", Models: map[string]Model{
		"a": {Name: "a", ToolCall: false},
	}}
	if hasToolCallModel(noTC) {
		t.Error("hasToolCallModel(noTC) = true, want false")
	}
	if hasToolCallModel(Provider{ID: "p"}) {
		t.Error("hasToolCallModel(empty) = true, want false")
	}
}

// TestFilterModelsForSDD asserts only tool_call models are kept, sorted by Name.
func TestFilterModelsForSDD(t *testing.T) {
	p := Provider{ID: "p", Models: map[string]Model{
		"z":   {Name: "Zeta", ToolCall: true},
		"a":   {Name: "Alpha", ToolCall: true},
		"mid": {Name: "Mid", ToolCall: false},
	}}
	got := FilterModelsForSDD(p)
	if len(got) != 2 {
		t.Fatalf("FilterModelsForSDD len = %d, want 2 (%v)", len(got), got)
	}
	if got[0].Name != "Alpha" || got[1].Name != "Zeta" {
		t.Errorf("FilterModelsForSDD = [%q, %q], want [Alpha, Zeta]", got[0].Name, got[1].Name)
	}
	for _, m := range got {
		if m.Name == "Mid" {
			t.Error("non-tool_call model Mid leaked into result")
		}
	}
}

// TestDetectAvailableProviders asserts sorted IDs, tool_call providers included,
// no-tool_call non-opencode excluded, and opencode always included if present.
func TestDetectAvailableProviders(t *testing.T) {
	t.Run("tool_call providers sorted", func(t *testing.T) {
		providers := map[string]Provider{
			"zeta":     {ID: "zeta", Models: map[string]Model{"m": {ToolCall: true}}},
			"requesty": {ID: "requesty", Models: map[string]Model{"m": {ToolCall: true}}},
		}
		got := DetectAvailableProviders(providers)
		want := []string{"requesty", "zeta"}
		if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
			t.Errorf("DetectAvailableProviders = %v, want %v", got, want)
		}
	})

	t.Run("opencode always included, no-tool_call excluded", func(t *testing.T) {
		providers := map[string]Provider{
			"opencode": {ID: "opencode", Models: map[string]Model{"m": {ToolCall: false}}},
			"foo":      {ID: "foo", Models: map[string]Model{"m": {ToolCall: false}}},
		}
		got := DetectAvailableProviders(providers)
		if len(got) != 1 || got[0] != "opencode" {
			t.Errorf("DetectAvailableProviders = %v, want [opencode]", got)
		}
	})
}
