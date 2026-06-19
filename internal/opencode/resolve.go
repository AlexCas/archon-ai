package opencode

import (
	"regexp"
	"sort"
	"strings"
)

// staticModelMap maps bare model names (from config.KnownModels) to their
// fully-qualified opencode provider/model IDs. This covers every key in
// config.KnownModels so init works without an opencode cache file.
var staticModelMap = map[string]string{
	"gpt-4":            "openai/gpt-4",
	"gpt-4o":           "openai/gpt-4o",
	"gpt-4o-mini":      "openai/gpt-4o-mini",
	"o3":               "openai/o3",
	"o3-mini":          "openai/o3-mini",
	"o4-mini":          "openai/o4-mini",
	"claude-sonnet-4":  "anthropic/claude-sonnet-4",
	"claude-haiku-4":   "anthropic/claude-haiku-4",
	"gemini-2.5-pro":   "google/gemini-2.5-pro",
	"gemini-2.5-flash": "google/gemini-2.5-flash",
}

// Resolve converts a bare or qualified model name to the provider/model form
// that opencode expects. It uses the default CachePath() for cache lookup.
// Resolution order:
//  1. Value already contains "/" → pass through verbatim (ok=true).
//  2. Name is in the static map → the static-map provider is canonical and wins
//     deterministically, regardless of what other cache providers may offer.
//     Within that provider, the cache is consulted for a more specific model ID
//     (e.g. a date-suffixed ID like "claude-sonnet-4-20250514"); if found, that
//     specific ID is returned. Otherwise the static-map qualified value is used.
//  3. Name is not in the static map but unambiguously present under exactly one
//     cache provider → use that provider/id pair. Providers are iterated in
//     sorted key order for determinism.
//  4. Returns ("", false) — caller should append a warning and skip the model field.
//
// The static-map tiebreak guarantees that known model names always resolve to
// their canonical provider, even when a real opencode cache lists the same bare
// name under an unrelated provider (e.g. neon/claude-sonnet-4).
func Resolve(name string) (qualified string, ok bool) {
	return resolveWithCachePath(name, CachePath())
}

// resolveWithCachePath is the internal implementation that accepts an explicit
// cache path. Used by Apply (which receives a CachePath from ApplyOptions) and
// by tests that need to control the cache location.
func resolveWithCachePath(name, cachePath string) (qualified string, ok bool) {
	if name == "" {
		return "", false
	}

	// 1. Already qualified.
	if strings.Contains(name, "/") {
		return name, true
	}

	// 2. Static map wins for known model names (deterministic tiebreak).
	//    The cache may contain the bare name under multiple providers; the static
	//    map resolves ambiguity by declaring the canonical provider.
	if staticQualified, found := staticModelMap[name]; found {
		// Extract the canonical provider from the static map value.
		staticProvider := ""
		if idx := strings.Index(staticQualified, "/"); idx >= 0 {
			staticProvider = staticQualified[:idx]
		}

		// Check the cache within the static-map provider for a more specific ID.
		// Prefer an exact match; otherwise accept only date-snapshot IDs of the
		// form "<name>-YYYYMMDD" (8 digits). Among multiple date snapshots pick
		// the newest (lexicographically last date string). This prevents a
		// distinct model like "claude-sonnet-4-5" from winning over the exact
		// bare name or a date snapshot "claude-sonnet-4-20250514".
		if staticProvider != "" {
			providers := LoadModelsOrEmpty(cachePath)
			if p, pOk := providers[staticProvider]; pOk {
				dateSnapshotRe := regexp.MustCompile(`^` + regexp.QuoteMeta(name) + `-\d{8}$`)
				exactMatch := ""
				var dateSnapshots []string
				for _, m := range p.Models {
					switch {
					case m.ID == name:
						exactMatch = m.ID
					case dateSnapshotRe.MatchString(m.ID):
						dateSnapshots = append(dateSnapshots, m.ID)
					}
				}
				if exactMatch != "" {
					return staticProvider + "/" + exactMatch, true
				}
				if len(dateSnapshots) > 0 {
					sort.Strings(dateSnapshots)
					// Pick the newest (highest date = last lexicographically).
					return staticProvider + "/" + dateSnapshots[len(dateSnapshots)-1], true
				}
			}
		}

		// Cache absent or static provider not listed: use the static map value.
		return staticQualified, true
	}

	// 3. Name not in static map: search all providers in sorted order for
	//    determinism (avoids random map iteration bias).
	providers := LoadModelsOrEmpty(cachePath)
	if len(providers) > 0 {
		providerIDs := make([]string, 0, len(providers))
		for id := range providers {
			providerIDs = append(providerIDs, id)
		}
		sort.Strings(providerIDs)

		for _, provID := range providerIDs {
			p := providers[provID]
			for _, m := range p.Models {
				if m.ID == name {
					return provID + "/" + m.ID, true
				}
			}
		}
	}

	// 4. Unresolvable.
	return "", false
}
