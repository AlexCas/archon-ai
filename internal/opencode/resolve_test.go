package opencode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// writeCacheFile writes a synthetic providers map as models.json into cacheDir.
func writeCacheFile(t *testing.T, cacheDir string, cache map[string]any) {
	t.Helper()
	data, err := json.Marshal(cache)
	if err != nil {
		t.Fatalf("marshal cache: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, "models.json"), data, 0o644); err != nil {
		t.Fatalf("write cache: %v", err)
	}
}

// multiProviderCache returns a synthetic cache where "claude-sonnet-4" appears
// under several providers (including two decoys) and the canonical "anthropic"
// provider lists a date-suffixed ID only — not the exact bare name.
// This exercises the tiebreak path: static-map provider wins, and the more
// specific ID is returned over the bare name.
func multiProviderCache() map[string]any {
	return map[string]any{
		// Decoy 1: neon lists the bare name exactly.
		"neon": map[string]any{
			"id":   "neon",
			"name": "Neon",
			"models": map[string]any{
				"claude-sonnet-4": map[string]any{
					"id":   "claude-sonnet-4",
					"name": "Claude Sonnet 4 (Neon)",
				},
			},
		},
		// Decoy 2: azure also lists the bare name.
		"azure": map[string]any{
			"id":   "azure",
			"name": "Azure",
			"models": map[string]any{
				"claude-sonnet-4": map[string]any{
					"id":   "claude-sonnet-4",
					"name": "Claude Sonnet 4 (Azure)",
				},
			},
		},
		// Canonical provider: anthropic uses the real date-suffixed ID.
		"anthropic": map[string]any{
			"id":   "anthropic",
			"name": "Anthropic",
			"models": map[string]any{
				"claude-sonnet-4-20250514": map[string]any{
					"id":   "claude-sonnet-4-20250514",
					"name": "Claude Sonnet 4 (2025-05-14)",
				},
			},
		},
	}
}

func TestResolve(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantQualified string
		wantOk        bool
		setupCache    func(t *testing.T) string // returns cache dir (containing models.json); empty = no cache
	}{
		{
			name:          "qualified ID passthrough",
			input:         "anthropic/claude-sonnet-4",
			wantQualified: "anthropic/claude-sonnet-4",
			wantOk:        true,
		},
		{
			name:          "cache hit",
			input:         "my-custom-model",
			wantQualified: "myprovider/my-custom-model",
			wantOk:        true,
			setupCache: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeCacheFile(t, dir, map[string]any{
					"myprovider": map[string]any{
						"id":   "myprovider",
						"name": "My Provider",
						"models": map[string]any{
							"my-custom-model": map[string]any{
								"id":   "my-custom-model",
								"name": "My Custom Model",
							},
						},
					},
				})
				return dir
			},
		},
		{
			name:          "cache miss static map fallback",
			input:         "gpt-4o",
			wantQualified: "openai/gpt-4o",
			wantOk:        true,
			setupCache: func(t *testing.T) string {
				t.Helper()
				// Return an empty temp dir (no models.json) to ensure the static map
				// path is exercised rather than a real cache hit.
				return t.TempDir()
			},
		},
		{
			name:          "unknown name returns not-ok",
			input:         "not-a-real-model",
			wantQualified: "",
			wantOk:        false,
			setupCache: func(t *testing.T) string {
				t.Helper()
				// Use empty cache dir so the static map path is reached and
				// the unknown name returns ok=false.
				return t.TempDir()
			},
		},
		// Tiebreak: static-map provider wins even when decoy providers list the
		// exact bare name.  The cache has anthropic listing a date-suffixed ID
		// only; the tiebreak must return anthropic's specific ID, not any decoy.
		{
			name:          "static-map provider wins over decoys (multi-provider cache)",
			input:         "claude-sonnet-4",
			wantQualified: "anthropic/claude-sonnet-4-20250514",
			wantOk:        true,
			setupCache: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeCacheFile(t, dir, multiProviderCache())
				return dir
			},
		},
		// Tiebreak with no cache entry for the static provider: fall back to the
		// static map value directly (no cache hit for anthropic).
		{
			name:          "static-map provider wins when canonical provider absent from cache",
			input:         "claude-sonnet-4",
			wantQualified: "anthropic/claude-sonnet-4",
			wantOk:        true,
			setupCache: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				// Cache has neon and azure but NOT anthropic.
				writeCacheFile(t, dir, map[string]any{
					"neon": map[string]any{
						"id":   "neon",
						"name": "Neon",
						"models": map[string]any{
							"claude-sonnet-4": map[string]any{
								"id":   "claude-sonnet-4",
								"name": "Claude Sonnet 4 (Neon)",
							},
						},
					},
					"azure": map[string]any{
						"id":   "azure",
						"name": "Azure",
						"models": map[string]any{
							"claude-sonnet-4": map[string]any{
								"id":   "claude-sonnet-4",
								"name": "Claude Sonnet 4 (Azure)",
							},
						},
					},
				})
				return dir
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cachePath := ""
			if tt.setupCache != nil {
				cacheDir := tt.setupCache(t)
				cachePath = filepath.Join(cacheDir, "models.json")
			}

			gotQualified, gotOk := resolveWithCachePath(tt.input, cachePath)
			if gotOk != tt.wantOk {
				t.Errorf("resolveWithCachePath(%q) ok = %v, want %v", tt.input, gotOk, tt.wantOk)
			}
			if gotQualified != tt.wantQualified {
				t.Errorf("resolveWithCachePath(%q) = %q, want %q", tt.input, gotQualified, tt.wantQualified)
			}
		})
	}
}

// TestResolve_StaticMapTiebreak_Deterministic verifies that resolving a known
// model name from a multi-provider cache is deterministic across many calls.
// Map iteration order in Go is randomised; this test catches any remaining
// reliance on it by running the resolution 100 times and checking consistency.
func TestResolve_StaticMapTiebreak_Deterministic(t *testing.T) {
	dir := t.TempDir()
	writeCacheFile(t, dir, multiProviderCache())
	cachePath := filepath.Join(dir, "models.json")

	const iterations = 100
	const input = "claude-sonnet-4"
	const want = "anthropic/claude-sonnet-4-20250514"

	for i := 0; i < iterations; i++ {
		got, ok := resolveWithCachePath(input, cachePath)
		if !ok {
			t.Fatalf("iteration %d: resolveWithCachePath(%q) ok=false", i, input)
		}
		if got != want {
			t.Fatalf("iteration %d: resolveWithCachePath(%q) = %q, want %q (non-deterministic!)", i, input, got, want)
		}
	}
}

// TestResolve_Fix4_DateSnapshotNewest verifies Fix 4: when the cache has both a
// decoy model (claude-sonnet-4-5) that is a distinct newer model and two date
// snapshots (claude-sonnet-4-20250514, claude-sonnet-4-20251022), the decoy must
// NOT win and the newest date snapshot must be returned.
func TestResolve_Fix4_DateSnapshotNewest(t *testing.T) {
	dir := t.TempDir()
	writeCacheFile(t, dir, map[string]any{
		"anthropic": map[string]any{
			"id":   "anthropic",
			"name": "Anthropic",
			"models": map[string]any{
				// Decoy: a distinct model that shares the prefix but is NOT a date snapshot.
				"claude-sonnet-4-5": map[string]any{
					"id":   "claude-sonnet-4-5",
					"name": "Claude Sonnet 4.5",
				},
				// Two date snapshots — the newest must win.
				"claude-sonnet-4-20250514": map[string]any{
					"id":   "claude-sonnet-4-20250514",
					"name": "Claude Sonnet 4 (2025-05-14)",
				},
				"claude-sonnet-4-20251022": map[string]any{
					"id":   "claude-sonnet-4-20251022",
					"name": "Claude Sonnet 4 (2025-10-22)",
				},
			},
		},
	})
	cachePath := filepath.Join(dir, "models.json")

	got, ok := resolveWithCachePath("claude-sonnet-4", cachePath)
	if !ok {
		t.Fatal("resolveWithCachePath returned ok=false")
	}
	const want = "anthropic/claude-sonnet-4-20251022"
	if got != want {
		t.Errorf("resolveWithCachePath(%q) = %q, want %q (decoy claude-sonnet-4-5 must not win; newest date snapshot must be selected)", "claude-sonnet-4", got, want)
	}
}

// TestResolve_Fix4_Gpt4DoesNotMatchGpt4Turbo verifies Fix 4: "gpt-4" must not
// resolve to "gpt-4-turbo" since "-turbo" is not 8 digits.
func TestResolve_Fix4_Gpt4DoesNotMatchGpt4Turbo(t *testing.T) {
	dir := t.TempDir()
	writeCacheFile(t, dir, map[string]any{
		"openai": map[string]any{
			"id":   "openai",
			"name": "OpenAI",
			"models": map[string]any{
				// Decoy: gpt-4-turbo must NOT match when resolving "gpt-4".
				"gpt-4-turbo": map[string]any{
					"id":   "gpt-4-turbo",
					"name": "GPT-4 Turbo",
				},
				// The exact bare name IS present — it should win.
				"gpt-4": map[string]any{
					"id":   "gpt-4",
					"name": "GPT-4",
				},
			},
		},
	})
	cachePath := filepath.Join(dir, "models.json")

	got, ok := resolveWithCachePath("gpt-4", cachePath)
	if !ok {
		t.Fatal("resolveWithCachePath returned ok=false")
	}
	const want = "openai/gpt-4"
	if got != want {
		t.Errorf("resolveWithCachePath(%q) = %q, want %q (gpt-4-turbo must not win)", "gpt-4", got, want)
	}
}

// TestResolve_Fix4_Gpt4WithoutExactMatch falls back to static map when neither
// an exact match nor a date snapshot is available for "gpt-4".
func TestResolve_Fix4_Gpt4WithoutExactMatch(t *testing.T) {
	dir := t.TempDir()
	// Cache has gpt-4-turbo only (not gpt-4 and no date-snapshot). Resolution
	// must fall through to the static map and return "openai/gpt-4".
	writeCacheFile(t, dir, map[string]any{
		"openai": map[string]any{
			"id":   "openai",
			"name": "OpenAI",
			"models": map[string]any{
				"gpt-4-turbo": map[string]any{
					"id":   "gpt-4-turbo",
					"name": "GPT-4 Turbo",
				},
			},
		},
	})
	cachePath := filepath.Join(dir, "models.json")

	got, ok := resolveWithCachePath("gpt-4", cachePath)
	if !ok {
		t.Fatal("resolveWithCachePath returned ok=false")
	}
	const want = "openai/gpt-4"
	if got != want {
		t.Errorf("resolveWithCachePath(%q) = %q, want %q (should fall back to static map when no exact/date-snapshot match)", "gpt-4", got, want)
	}
}

// TestResolve_NonStaticName_SortedProviderOrder verifies that for a model name
// that is NOT in the static map and appears under multiple providers, the
// resolution picks the lexicographically first provider (sorted), not an
// arbitrary one from random map iteration.
func TestResolve_NonStaticName_SortedProviderOrder(t *testing.T) {
	dir := t.TempDir()
	writeCacheFile(t, dir, map[string]any{
		"zzz-last": map[string]any{
			"id":   "zzz-last",
			"name": "ZZZ Provider",
			"models": map[string]any{
				"my-exotic-model": map[string]any{
					"id":   "my-exotic-model",
					"name": "Exotic Model",
				},
			},
		},
		"aaa-first": map[string]any{
			"id":   "aaa-first",
			"name": "AAA Provider",
			"models": map[string]any{
				"my-exotic-model": map[string]any{
					"id":   "my-exotic-model",
					"name": "Exotic Model",
				},
			},
		},
	})
	cachePath := filepath.Join(dir, "models.json")

	const iterations = 100
	const input = "my-exotic-model"
	const want = "aaa-first/my-exotic-model"

	for i := 0; i < iterations; i++ {
		got, ok := resolveWithCachePath(input, cachePath)
		if !ok {
			t.Fatalf("iteration %d: resolveWithCachePath(%q) ok=false", i, input)
		}
		if got != want {
			t.Fatalf("iteration %d: resolveWithCachePath(%q) = %q, want %q (non-deterministic!)", i, input, got, want)
		}
	}
}
