package models

import (
	"context"
	"time"

	"github.com/archon-ai/archon/internal/config"
)

// listTimeout bounds live opencode enumeration so a slow or hung CLI never
// blocks the caller.
const listTimeout = 2 * time.Second

// ResolveModels composes the ordered catalog of offered models from the injected
// detector and opencode lister. Only catalogs whose CLI is detected are
// appended; absent CLIs are skipped (hidden, never curated). For opencode the
// resolver prefers the live enumeration and falls back to the curated
// config.OpencodeModels only when enumeration fails or returns nothing. detect()
// is called once and the resulting map reused.
//
// The structure is catalog-agnostic in spirit: claude and opencode are the only
// catalogs on this branch, but adding future curated catalogs (e.g. Gemini /
// OpenAI) is a matter of gating each on its detected CLI the same way.
func ResolveModels(detect CLIDetector, lister OpencodeLister) []string {
	present := detect()
	var out []string

	if present["claude"] {
		out = append(out, config.ClaudeModels...)
	}

	if present["opencode"] {
		ctx, cancel := context.WithTimeout(context.Background(), listTimeout)
		defer cancel()
		if live, err := lister.List(ctx); err == nil && len(live) > 0 {
			out = append(out, live...)
		} else {
			out = append(out, config.OpencodeModels...) // installed-but-failed fallback
		}
	}

	// Future curated catalogs gate the same way when they exist.
	return out
}

// Resolve is the default convenience used by the TUI: it wires the real
// PATH detector and the real opencode lister.
func Resolve() []string {
	return ResolveModels(LookPathDetector, execLister{})
}
