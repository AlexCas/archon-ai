package config

import (
	"fmt"
	"strings"
	"unicode"
)

type ModelConfig struct {
	Default string            `yaml:"default,omitempty"`
	Phases  map[string]string `yaml:"phases,omitempty"`
}

// ClaudeModels is the curated, ordered list of Claude models offered as static
// choices in the TUI. The list is intentionally small and current; users who
// need another model can always type a free-form string instead.
var ClaudeModels = []string{
	"claude-opus-4-8",
	"claude-sonnet-4-6",
	"claude-haiku-4-5",
}

// OpencodeModels is the curated, ordered list of models available through the
// Opencode Go subscription, offered as static choices in the TUI. Adjust this
// list as the subscription catalog changes; the free-form string option remains
// available for anything not listed here.
var OpencodeModels = []string{
	"deepseek-v4-flash",
	"deepseek-v4-pro",
	"glm-5",
	"glm-5.1",
	"kimi-k2.5",
	"kimi-k2.6",
	"qwen3.6-plus",
	"qwen3.7-plus",
}

// StaticModels returns the full ordered list of statically-selectable models,
// Claude first then Opencode Go. The TUI renders these as pickable options and
// still allows a free-form entry.
func StaticModels() []string {
	out := make([]string, 0, len(ClaudeModels)+len(OpencodeModels))
	out = append(out, ClaudeModels...)
	out = append(out, OpencodeModels...)
	return out
}

// KnownModels is the validation set: any statically-listed model is "known".
// Validation is advisory only — unknown (free-form) models are accepted with a
// warning, never rejected.
var KnownModels = func() map[string]bool {
	m := make(map[string]bool)
	for _, name := range StaticModels() {
		m[name] = true
	}
	return m
}()

var ValidPhases = map[string]bool{
	"explore": true,
	"propose": true,
	"spec":    true,
	"design":  true,
	"tasks":   true,
	"apply":   true,
	"verify":  true,
	"archive": true,
}

// PhaseOrder is the canonical, delegated SDD phase order. It excludes judge,
// which is not delegated to an sdd-* sub-agent. Iterating this slice (rather
// than a map) gives deterministic, byte-identical output across runs.
var PhaseOrder = []string{"explore", "propose", "spec", "design", "tasks", "apply", "verify", "archive"}

// claudeFamilies are the Claude model family aliases the delegation tool
// accepts. NormalizeModel collapses display strings and full IDs down to one
// of these.
var claudeFamilies = []string{"opus", "sonnet", "haiku"}

// PhaseModel pairs an SDD phase with its resolved, normalized model alias.
type PhaseModel struct {
	Phase string
	Model string
}

// NormalizeModel maps a configured/display model value to an alias the
// delegation tool accepts (opus|sonnet|haiku). It is case-insensitive,
// tolerant of extra version digits, and idempotent for values already in
// canonical/accepted form. ok is false when no known Claude model resolves
// (e.g. typos like "Opues 4.8", or non-Claude Opencode models like "glm-5").
//
// A family matches only as a whole token: the value is split on non-alphanumeric
// boundaries, so "claude-opus-4-8", "opus 4.8" and "opus" all resolve to "opus",
// while a word that merely contains a family substring (e.g. "octopus") does
// not. When a value names more than one family the fixed priority order
// opus→sonnet→haiku wins, keeping resolution deterministic.
func NormalizeModel(s string) (id string, ok bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "", false
	}
	tokens := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	for _, fam := range claudeFamilies {
		for _, tok := range tokens {
			if tok == fam {
				return fam, true
			}
		}
	}
	return "", false
}

// ResolvePhaseModels returns phase→alias pairs in canonical PhaseOrder,
// omitting any phase that resolves to nothing. For each phase it tries the
// explicit Phases entry, falls back to Default, and omits the line if neither
// normalizes. The function is pure and never mutates mc.
func ResolvePhaseModels(mc ModelConfig) []PhaseModel {
	var out []PhaseModel
	for _, p := range PhaseOrder {
		id, ok := NormalizeModel(mc.Phases[p])
		if !ok {
			id, ok = NormalizeModel(mc.Default)
		}
		if !ok {
			continue
		}
		out = append(out, PhaseModel{Phase: p, Model: id})
	}
	return out
}

func Validate(model string) string {
	if model == "" {
		return ""
	}
	if KnownModels[model] {
		return ""
	}
	if _, ok := NormalizeModel(model); ok {
		return ""
	}
	return fmt.Sprintf("warning: %q is not a known model (accepted anyway)", model)
}
