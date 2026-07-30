package config

import (
	"fmt"
	"io"
	"net/url"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

// ModelRef is a structured provider+model assignment. An empty Provider is a
// valid advisory-only state (FullID returns the bare Model); a bare legacy alias
// decodes to this state and re-marshals as the same scalar.
type ModelRef struct {
	Provider string `yaml:"provider,omitempty"`
	Model    string `yaml:"model,omitempty"`
	Effort   string `yaml:"effort,omitempty"`
	// BaseURL is an optional OpenAI-compatible endpoint (e.g. a local Ollama or
	// LocalAI server). Empty for every remote/hosted provider. Only the
	// OpenCode generation path honors it (see internal/initcmd/opencode_mode.go);
	// the Claude path warns and ignores it.
	BaseURL string `yaml:"base_url,omitempty"`
}

// FullID returns the provider-qualified id used by delegations:
//   - Model already contains "/" -> returned as-is (no double-prefix).
//   - non-empty Provider + bare Model -> "<provider>/<model>".
//   - empty Provider -> the bare Model (no leading slash).
func (r ModelRef) FullID() string {
	if strings.Contains(r.Model, "/") {
		return r.Model
	}
	if r.Provider == "" {
		return r.Model
	}
	return r.Provider + "/" + r.Model
}

// UnmarshalYAML accepts either a legacy scalar ("provider/model" or bare "model")
// or a structured mapping ({provider, model, effort}).
func (r *ModelRef) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		s := node.Value
		if i := strings.Index(s, "/"); i >= 0 { // split on the FIRST "/"
			r.Provider = s[:i]
			r.Model = s[i+1:]
			return nil
		}
		r.Provider = "" // bare alias -> empty provider, NEVER guessed
		r.Model = s
		return nil
	}
	type modelRefAlias ModelRef // strips the custom method -> no recursion
	var tmp modelRefAlias
	if err := node.Decode(&tmp); err != nil {
		return err
	}
	*r = ModelRef(tmp)
	return nil
}

// MarshalYAML emits a scalar equal to FullID() when Effort == "" && BaseURL ==
// "", keeping unmigrated flat-string configs byte-identical — both a bare
// alias ("opus") and a provider-qualified scalar ("anthropic/claude-...")
// re-serialize to the same one-line form they were loaded from. A ref that
// carries an Effort and/or a BaseURL (neither has a legacy scalar
// representation) marshals as a mapping.
func (r ModelRef) MarshalYAML() (any, error) {
	if r.Effort == "" && r.BaseURL == "" {
		return r.FullID(), nil // SCALAR — byte-identical for unmigrated configs
	}
	type modelRefAlias ModelRef // mapping, no recursion
	return modelRefAlias(r), nil
}

// ValidateBaseURL is an advisory-only check: it never returns an error and
// never blocks the caller. It warns to w when BaseURL is set but Provider is
// empty (local routing needs a provider id), and when BaseURL does not parse
// as an http/https URL with a non-empty host. A ref with an empty BaseURL is
// always silent — there is nothing to validate.
func ValidateBaseURL(ref ModelRef, w io.Writer) {
	if ref.BaseURL == "" {
		return
	}
	if ref.Provider == "" {
		fmt.Fprintln(w, "warning: base_url is set but provider is empty — provider id required for local model routing")
	}
	if !isValidHTTPURL(ref.BaseURL) {
		fmt.Fprintf(w, "warning: base_url %q is not a valid http/https URL\n", ref.BaseURL)
	}
}

// isValidHTTPURL reports whether s parses as an absolute http or https URL
// with a non-empty host.
func isValidHTTPURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return u.Host != ""
}

// ParseModelRef splits a user-supplied "provider/model" string into a ModelRef.
// A bare value (no "/") yields an empty Provider (advisory-only). Splitting on
// the FIRST "/" mirrors FullID/UnmarshalYAML, so ParseModelRef and UnmarshalYAML agree.
func ParseModelRef(s string) ModelRef {
	if i := strings.Index(s, "/"); i >= 0 {
		return ModelRef{Provider: s[:i], Model: s[i+1:]}
	}
	return ModelRef{Model: s}
}

type ModelConfig struct {
	Default ModelRef            `yaml:"default,omitempty"`
	Leader  ModelRef            `yaml:"leader,omitempty"`
	Phases  map[string]ModelRef `yaml:"phases,omitempty"`
}

// HasAny reports whether any ref carries a model id OR a base_url. It is the
// single source of truth for the "(none configured)" guard on both the
// `config list` and `status` surfaces (Invariant 4).
func (mc ModelConfig) HasAny() bool {
	if mc.Default.FullID() != "" || mc.Default.BaseURL != "" {
		return true
	}
	if mc.Leader.FullID() != "" || mc.Leader.BaseURL != "" {
		return true
	}
	for _, r := range mc.Phases {
		if r.FullID() != "" || r.BaseURL != "" {
			return true
		}
	}
	return false
}

// ClaudeModels is the curated, ordered list of Claude models offered as static
// choices in the TUI. The list is intentionally small and current; users who
// need another model can always type a free-form string instead.
var ClaudeModels = []string{
	"claude-opus-4-8",
	"claude-sonnet-4-6",
	"claude-haiku-4-5",
}

// GeminiModels is the curated, ordered list of Google Gemini models offered as
// static choices in the TUI. The list is intentionally small and current; users
// who need another model can always type a free-form string instead.
var GeminiModels = []string{
	"gemini-2.5-pro",
	"gemini-2.5-flash",
	"gemini-2.0-flash",
}

// OpenAIModels is the curated, ordered list of OpenAI models offered as static
// choices in the TUI. The list is intentionally small and current; users who
// need another model can always type a free-form string instead.
var OpenAIModels = []string{
	"gpt-4o",
	"gpt-4o-mini",
	"gpt-4.1",
	"o3",
	"o4-mini",
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

// StaticModels returns the full ordered list of statically-selectable models in
// fixed provider precedence: Claude → Gemini → OpenAI → Opencode Go. The TUI
// renders these as pickable options and still allows a free-form entry.
func StaticModels() []string {
	out := make([]string, 0, len(ClaudeModels)+len(GeminiModels)+len(OpenAIModels)+len(OpencodeModels))
	out = append(out, ClaudeModels...)
	out = append(out, GeminiModels...)
	out = append(out, OpenAIModels...)
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
	"judge":   true,
	"archive": true,
}

// PhaseOrder is the canonical, delegated SDD phase order. Iterating this slice
// (rather than a map) gives deterministic, byte-identical output across runs.
var PhaseOrder = []string{"explore", "propose", "spec", "design", "tasks", "apply", "verify", "judge", "archive"}

// claudeFamilies are the Claude model family aliases the delegation tool
// accepts. NormalizeModel collapses display strings and full IDs down to one
// of these.
var claudeFamilies = []string{"opus", "sonnet", "haiku"}

// providerFamily describes one provider row for NormalizeModel. A row is either
// family-based (families non-empty, used by Claude for whole-token alias
// collapsing) or catalog-based (catalog non-empty, matched by exact
// case-insensitive id and emitted as-is). Exactly one field is populated per
// row.
type providerFamily struct {
	families []string
	catalog  []string
}

// providerFamilies is the ordered cross-provider precedence table NormalizeModel
// walks. The order is fixed — Claude → Gemini → OpenAI → Opencode — so a value
// that could match more than one provider resolves deterministically to the
// earliest row.
var providerFamilies = []providerFamily{
	{families: claudeFamilies},
	{catalog: GeminiModels},
	{catalog: OpenAIModels},
	{catalog: OpencodeModels},
}

// PhaseModel pairs an SDD phase with its resolved, normalized model alias.
type PhaseModel struct {
	Phase    string
	Model    string
	Provider string // resolved ref's raw Provider id; "" when the ref has none
	Effort   string // resolved ModelRef.Effort (variant); "" = provider default
	BaseURL  string // resolved ref's BaseURL; "" = no local endpoint
}

// NormalizeModel maps a configured/display model value to the canonical
// identifier a provider accepts. It walks the fixed provider precedence
// Claude → Gemini → OpenAI → Opencode and returns the first match; the per-row
// match rule and canonical output differ by provider:
//
//   - Claude (family row): the value is matched as a whole token and collapsed
//     to a family alias (opus|sonnet|haiku). The value is split on
//     non-alphanumeric boundaries, so "claude-opus-4-8", "opus 4.8" and "opus"
//     all resolve to "opus", while a word that merely contains a family
//     substring (e.g. "octopus") does not. When a value names more than one
//     family the fixed priority opus→sonnet→haiku wins.
//   - Gemini/OpenAI/Opencode (catalog rows): the trimmed, lowercased value must
//     equal a curated catalog id exactly (case-insensitive); the canonical id is
//     emitted as-is, e.g. "gpt-4o" → "gpt-4o", "glm-5" → "glm-5".
//
// Matching is case-insensitive and idempotent for values already in
// canonical/accepted form. ok is false when no provider row resolves (e.g. typos
// like "Opues 4.8", or a model not in any curated catalog).
func NormalizeModel(s string) (id string, ok bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "", false
	}
	tokens := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	for _, pf := range providerFamilies {
		// Claude-style family row: whole-token alias match.
		for _, fam := range pf.families {
			for _, tok := range tokens {
				if tok == fam {
					return fam, true
				}
			}
		}
		// Catalog row: exact case-insensitive id match, emitted as-is.
		for _, id := range pf.catalog {
			if s == strings.ToLower(id) {
				return id, true
			}
		}
	}
	return "", false
}

// ResolvePhaseModels returns phase->model pairs in canonical PhaseOrder. For each
// phase it prefers the explicit Phases entry, falls back to Default, and omits the
// phase when neither yields a model. The emitted Model is ref.FullID():
// "<provider>/<model>" when a provider is present, else the bare alias.
func ResolvePhaseModels(mc ModelConfig) []PhaseModel {
	var out []PhaseModel
	for _, p := range PhaseOrder {
		ref, ok := mc.Phases[p]
		if !ok || ref.Model == "" {
			ref = mc.Default
		}
		if ref.Model == "" {
			continue
		}
		out = append(out, PhaseModel{Phase: p, Model: ref.FullID(), Provider: ref.Provider, Effort: ref.Effort, BaseURL: ref.BaseURL})
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
