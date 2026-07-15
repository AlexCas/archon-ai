package models

import (
	"context"
	"errors"
	"testing"

	"github.com/archon-ai/archon/internal/config"
	"github.com/archon-ai/archon/internal/opencode"
)

// fakeLister is an injectable OpencodeLister for unit tests; it never touches a
// real subprocess.
type fakeLister struct {
	models []string
	err    error
	called bool
}

func (f *fakeLister) List(ctx context.Context) ([]string, error) {
	f.called = true
	return f.models, f.err
}

// emptyCache is a CacheReader stub that returns no providers.
func emptyCache() (map[string]opencode.Provider, error) {
	return nil, nil
}

func contains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

func containsAny(list, candidates []string) bool {
	for _, c := range candidates {
		if contains(list, c) {
			return true
		}
	}
	return false
}

// Scenario: "Installed opencode shows the live catalog".
// S1d-2: third arg is the empty-cache stub so shell-out/curated behavior is preserved.
func TestResolveModels_InstalledOpencodeShowsLiveCatalog(t *testing.T) {
	detect := func() map[string]bool {
		return map[string]bool{"opencode": true, "claude": true}
	}
	// Use live values that are deliberately NOT in the curated OpencodeModels
	// list so we can assert the curated sentinels did not leak in.
	live := []string{"glm-9.9-live", "kimi-live-only"}
	lister := &fakeLister{models: live}

	out := ResolveModels(detect, lister, emptyCache)

	for _, m := range live {
		if !contains(out, m) {
			t.Errorf("expected live model %q in output %v", m, out)
		}
	}
	// The curated opencode sentinels must NOT appear when the live catalog is used.
	if containsAny(out, config.OpencodeModels) {
		t.Errorf("curated OpencodeModels leaked into output %v despite live catalog", out)
	}
	// A present agent's curated catalog (claude) still appears.
	for _, m := range config.ClaudeModels {
		if !contains(out, m) {
			t.Errorf("expected claude model %q in output %v", m, out)
		}
	}
}

// Scenario: "Only installed agents' models are offered".
// S1d-2: third arg is the empty-cache stub.
func TestResolveModels_OnlyInstalledAgentsOffered(t *testing.T) {
	detect := func() map[string]bool {
		return map[string]bool{"opencode": false, "claude": true}
	}
	// Lister would return data, but opencode is absent so it must never be used.
	lister := &fakeLister{models: []string{"glm-5.2"}}

	out := ResolveModels(detect, lister, emptyCache)

	// No opencode models — neither live nor curated — should appear.
	if contains(out, "glm-5.2") {
		t.Errorf("live opencode model leaked despite opencode absent: %v", out)
	}
	if containsAny(out, config.OpencodeModels) {
		t.Errorf("curated OpencodeModels appeared despite opencode absent: %v", out)
	}
	// The present agent's (claude) models remain.
	for _, m := range config.ClaudeModels {
		if !contains(out, m) {
			t.Errorf("expected claude model %q in output %v", m, out)
		}
	}
}

// Scenario: "Live enumeration error falls back silently".
// S1d-2: third arg is the empty-cache stub so shell-out/curated behavior is preserved.
func TestResolveModels_LiveEnumerationFallsBackToCurated(t *testing.T) {
	detect := func() map[string]bool {
		return map[string]bool{"opencode": true}
	}

	cases := []struct {
		name   string
		lister *fakeLister
	}{
		{"error from lister", &fakeLister{err: errors.New("boom")}},
		{"empty result", &fakeLister{models: []string{}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := ResolveModels(detect, tc.lister, emptyCache)
			if !equalStrings(out, config.OpencodeModels) {
				t.Errorf("expected curated fallback %v, got %v", config.OpencodeModels, out)
			}
		})
	}
}

// TestResolveModels_PrefersCache asserts the cache is used when non-empty and
// the shell-out lister is NOT invoked. (C4 prefers cache)
func TestResolveModels_PrefersCache(t *testing.T) {
	detect := func() map[string]bool {
		return map[string]bool{"opencode": true}
	}

	// Cache returns an opencode provider with bare keys
	cacheReader := func() (map[string]opencode.Provider, error) {
		return map[string]opencode.Provider{
			"opencode": {
				ID:   "opencode",
				Name: "Opencode",
				Models: map[string]opencode.Model{
					"deepseek-v4-pro": {ID: "deepseek-v4-pro", ToolCall: true},
					"glm-5":           {ID: "glm-5"},
				},
			},
		}, nil
	}

	// Sentinel lister that records whether it was called
	lister := &fakeLister{models: []string{"should-not-appear"}}

	out := ResolveModels(detect, lister, cacheReader)

	// Lister must NOT have been invoked
	if lister.called {
		t.Error("lister was invoked despite non-empty cache (cache should be preferred)")
	}

	// Output must contain FullID-form names from the cache
	if !contains(out, "opencode/deepseek-v4-pro") {
		t.Errorf("output %v missing opencode/deepseek-v4-pro", out)
	}
	if !contains(out, "opencode/glm-5") {
		t.Errorf("output %v missing opencode/glm-5", out)
	}

	// The sentinel live model must not appear
	if contains(out, "should-not-appear") {
		t.Errorf("lister sentinel leaked into output %v", out)
	}
}

// TestResolveModels_FallbackWhenCacheEmpty asserts that an empty cache triggers
// the shell-out fallback. (C4 fallback + PR #45 reachable)
func TestResolveModels_FallbackWhenCacheEmpty(t *testing.T) {
	detect := func() map[string]bool {
		return map[string]bool{"opencode": true}
	}

	// Cache returns empty map (simulates absent/empty cache)
	cacheReader := func() (map[string]opencode.Provider, error) {
		return map[string]opencode.Provider{}, nil
	}

	liveNames := []string{"glm-live-from-shell", "kimi-live-from-shell"}
	lister := &fakeLister{models: liveNames}

	out := ResolveModels(detect, lister, cacheReader)

	// Lister must have been invoked (shell-out fallback path)
	if !lister.called {
		t.Error("lister was not invoked for empty cache (shell-out fallback expected)")
	}

	for _, m := range liveNames {
		if !contains(out, m) {
			t.Errorf("output %v missing shell-out model %q", out, m)
		}
	}
}

// TestCacheModelNames_Mapping asserts FullID form for opencode bare keys, that
// the catalog is SCOPED to the opencode provider (non-opencode providers are
// excluded — they surface later via the provider→model picker), and that the
// output is sorted. (C4 FullID form, opencode-go scope to match the shell-out)
func TestCacheModelNames_Mapping(t *testing.T) {
	cacheReader := func() (map[string]opencode.Provider, error) {
		return map[string]opencode.Provider{
			"opencode": {
				ID:   "opencode",
				Name: "Opencode",
				Models: map[string]opencode.Model{
					"glm-5":           {ID: "glm-5"},
					"deepseek-v4-pro": {ID: "deepseek-v4-pro"},
				},
			},
			"requesty": {
				ID:   "requesty",
				Name: "Requesty",
				Models: map[string]opencode.Model{
					"xai/grok-4": {ID: "xai/grok-4"},
				},
			},
		}, nil
	}

	got := cacheModelNames(cacheReader)

	// bare opencode keys -> provider/key (FullID)
	if !contains(got, "opencode/glm-5") {
		t.Errorf("output %v missing opencode/glm-5", got)
	}
	if !contains(got, "opencode/deepseek-v4-pro") {
		t.Errorf("output %v missing opencode/deepseek-v4-pro", got)
	}
	// non-opencode providers are NOT in the flat catalog (scoped to opencode-go)
	if contains(got, "xai/grok-4") {
		t.Errorf("output %v leaked non-opencode model xai/grok-4", got)
	}
	if len(got) != 2 {
		t.Errorf("expected exactly 2 opencode models, got %d: %v", len(got), got)
	}

	// output must be sorted
	for i := 1; i < len(got); i++ {
		if got[i] < got[i-1] {
			t.Errorf("output not sorted: %v", got)
			break
		}
	}
}

// TestCacheModelNames_NoOpencodeProvider asserts that a cache with providers but
// no "opencode" entry yields nil (caller falls back to the shell-out).
func TestCacheModelNames_NoOpencodeProvider(t *testing.T) {
	cacheReader := func() (map[string]opencode.Provider, error) {
		return map[string]opencode.Provider{
			"requesty": {
				ID:     "requesty",
				Name:   "Requesty",
				Models: map[string]opencode.Model{"xai/grok-4": {ID: "xai/grok-4"}},
			},
		}, nil
	}
	if got := cacheModelNames(cacheReader); got != nil {
		t.Errorf("expected nil for cache without opencode provider, got %v", got)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
