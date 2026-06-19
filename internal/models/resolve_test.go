package models

import (
	"context"
	"errors"
	"testing"

	"github.com/archon-ai/archon/internal/config"
)

// fakeLister is an injectable OpencodeLister for unit tests; it never touches a
// real subprocess.
type fakeLister struct {
	models []string
	err    error
}

func (f fakeLister) List(ctx context.Context) ([]string, error) {
	return f.models, f.err
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
func TestResolveModels_InstalledOpencodeShowsLiveCatalog(t *testing.T) {
	detect := func() map[string]bool {
		return map[string]bool{"opencode": true, "claude": true}
	}
	// Use live values that are deliberately NOT in the curated OpencodeModels
	// list so we can assert the curated sentinels did not leak in.
	live := []string{"glm-9.9-live", "kimi-live-only"}
	lister := fakeLister{models: live}

	out := ResolveModels(detect, lister)

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
func TestResolveModels_OnlyInstalledAgentsOffered(t *testing.T) {
	detect := func() map[string]bool {
		return map[string]bool{"opencode": false, "claude": true}
	}
	// Lister would return data, but opencode is absent so it must never be used.
	lister := fakeLister{models: []string{"glm-5.2"}}

	out := ResolveModels(detect, lister)

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
func TestResolveModels_LiveEnumerationFallsBackToCurated(t *testing.T) {
	detect := func() map[string]bool {
		return map[string]bool{"opencode": true}
	}

	cases := []struct {
		name   string
		lister fakeLister
	}{
		{"error from lister", fakeLister{err: errors.New("boom")}},
		{"empty result", fakeLister{models: []string{}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := ResolveModels(detect, tc.lister)
			if !equalStrings(out, config.OpencodeModels) {
				t.Errorf("expected curated fallback %v, got %v", config.OpencodeModels, out)
			}
		})
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
