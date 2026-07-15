package models

import (
	"context"
	"sort"
	"time"

	"github.com/archon-ai/archon/internal/config"
	"github.com/archon-ai/archon/internal/opencode"
)

// listTimeout bounds live opencode enumeration so a slow or hung CLI never
// blocks the caller.
const listTimeout = 2 * time.Second

// CacheReader loads the opencode provider cache. The real implementation reads
// the default cache file; tests inject a fake returning canned providers.
type CacheReader func() (map[string]opencode.Provider, error)

// defaultCacheReader reads ~/.cache/opencode/models.json, empty-on-missing.
func defaultCacheReader() (map[string]opencode.Provider, error) {
	path, err := opencode.DefaultCachePath()
	if err != nil {
		return nil, err
	}
	return opencode.LoadModelsOrEmpty(path)
}

// opencodeProviderID is the cache provider key whose models archon offers in
// this slice — scoped to match the shell-out fallback (`opencode models
// opencode-go`). Other providers in the cache are surfaced later via the
// dedicated provider→model picker, not this flat catalog.
const opencodeProviderID = "opencode"

// cacheModelNames flattens the opencode-go provider's models from the cache into
// FullID strings (opencode/<key>). Returns nil when the cache is
// absent/empty/errored or has no opencode provider (caller falls back to the
// shell-out). Scoped to the opencode provider so the catalog matches the
// shell-out fallback's scope rather than every provider in the cache.
func cacheModelNames(cache CacheReader) []string {
	providers, err := cache()
	if err != nil || len(providers) == 0 {
		return nil
	}
	prov, ok := providers[opencodeProviderID]
	if !ok {
		return nil
	}
	var names []string
	for key := range prov.Models {
		ref := config.ModelRef{Provider: opencodeProviderID, Model: key}
		names = append(names, ref.FullID())
	}
	sort.Strings(names) // determinism: map iteration is unordered
	return names
}

// ResolveModels composes the ordered catalog of offered models from the injected
// detector, opencode lister, and cache reader. Only catalogs whose CLI is detected
// are appended; absent CLIs are skipped (hidden, never curated). For opencode the
// resolver prefers the cache (FullID form), falls back to the live shell-out
// (PR #45 path), and finally falls back to the curated config.OpencodeModels only
// when both the cache and shell-out return nothing. detect() is called once and
// the resulting map reused.
func ResolveModels(detect CLIDetector, lister OpencodeLister, cache CacheReader) []string {
	present := detect()
	var out []string

	if present["claude"] {
		out = append(out, config.ClaudeModels...)
	}

	if present["opencode"] {
		if names := cacheModelNames(cache); len(names) > 0 {
			out = append(out, names...) // cache FIRST, FullID form
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), listTimeout)
			defer cancel()
			if live, err := lister.List(ctx); err == nil && len(live) > 0 {
				out = append(out, live...) // PR #45 shell-out fallback
			} else {
				out = append(out, config.OpencodeModels...) // installed-but-failed fallback
			}
		}
	}

	// Future curated catalogs gate the same way when they exist.
	return out
}

// Resolve is the default convenience used by the TUI: it wires the real
// PATH detector, the real opencode lister, and the real cache reader.
func Resolve() []string {
	return ResolveModels(LookPathDetector, execLister{}, defaultCacheReader)
}
