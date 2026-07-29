package initcmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/archon-ai/archon/internal/config"
)

// leaderAgentName is the key under "agent" in opencode.json that holds the
// Archon SDD orchestration leader.
const leaderAgentName = "archon-leader"

// archonLeaderAgent is the fixed shape written under agent.archon-leader.
// Struct field declaration order is the deterministic JSON output order.
type archonLeaderAgent struct {
	Mode        string `json:"mode"`              // always "primary"
	Description string `json:"description"`       // fixed cosmetic label
	Model       string `json:"model"`             // verbatim models.leader
	Variant     string `json:"variant,omitempty"` // effort/reasoning level; omitted when empty
	Prompt      string `json:"prompt"`            // "{file:./AGENTS.md}"
}

// phaseAgentName returns the opencode.json agent key for an SDD phase, e.g.
// "archon-spec".
func phaseAgentName(phase string) string { return "archon-" + phase }

// archonPhaseAgent is the fixed shape written under agent.archon-<phase> for
// each resolvable SDD phase. Field declaration order is the deterministic JSON
// output order.
type archonPhaseAgent struct {
	Mode        string `json:"mode"`              // always "subagent"
	Hidden      bool   `json:"hidden"`            // always true (no omitempty)
	Model       string `json:"model"`             // resolved per-phase FullID
	Variant     string `json:"variant,omitempty"` // effort/reasoning level; omitted when empty
	Description string `json:"description"`       // "Archon SDD <phase> phase"
	Prompt      string `json:"prompt"`            // "{file:./AGENTS.md}"
}

// MergeOpencodeAgent is the exported integration seam for callers outside this
// package (notably the TUI save path). It delegates to mergeOpencodeAgent so
// init and the TUI share a single writer implementation and produce
// byte-identical opencode.json output.
//
// Caller audit (PR-A A1, local-model-provider): grepping every exported
// symbol in this package across the repo shows exactly two external callers —
// internal/tui/model.go (line ~372, passes io.Discard) and
// internal/tui/model_test.go (line ~766, passes io.Discard). init.go is the
// same-package caller of the unexported mergeOpencodeAgent and passes
// os.Stderr. No third caller exists, so widening this signature with an
// io.Writer param is safe.
func MergeOpencodeAgent(projectDir string, models config.ModelConfig, w io.Writer) (written string, err error) {
	return mergeOpencodeAgent(projectDir, models, w)
}

// mergeOpencodeAgent additively merges agent.archon-leader and one
// agent.archon-<phase> subagent per resolvable SDD phase into
// <projectDir>/opencode.json. It is a no-op (returns "", nil) when the leader
// FullID is empty and ResolvePhaseModels returns no phases.
//
// When one or more resolved refs (phases, plus the leader) carry a non-empty
// BaseURL, it also merges a top-level "provider" block built by
// buildProviderBlock; w receives any coalescing-conflict warnings.
//
// Existing top-level keys and existing agents/providers are preserved; only
// agent.archon-leader, agent.archon-<phase>, and archon-built provider.<id>
// entries are set. It never writes a default_agent key. The whole document is
// marshaled with json.MarshalIndent (sorted map keys) plus a trailing
// newline, then written atomically via a temp file + os.Rename — the same
// pattern as config.Save and writeTemplate — so re-runs yield byte-identical
// output. It returns the written path ("" when nothing was written) so the
// caller can register it for rollback.
func mergeOpencodeAgent(projectDir string, models config.ModelConfig, w io.Writer) (written string, err error) {
	leaderFull := models.Leader.FullID()
	phases := config.ResolvePhaseModels(models)
	if leaderFull == "" && len(phases) == 0 {
		return "", nil
	}

	path := filepath.Join(projectDir, "opencode.json")

	doc := map[string]any{}
	if data, readErr := os.ReadFile(path); readErr == nil {
		if err := json.Unmarshal(data, &doc); err != nil {
			return "", fmt.Errorf("parse opencode.json: %w", err)
		}
	} else if !os.IsNotExist(readErr) {
		return "", fmt.Errorf("read opencode.json: %w", readErr)
	}

	agents, ok := doc["agent"].(map[string]any)
	if !ok {
		agents = map[string]any{}
	}
	if leaderFull != "" {
		agents[leaderAgentName] = archonLeaderAgent{
			Mode:        "primary",
			Description: "Archon SDD orchestration leader",
			Model:       leaderFull,
			Variant:     models.Leader.Effort,
			Prompt:      "{file:./AGENTS.md}",
		}
	}
	for _, pm := range phases {
		agents[phaseAgentName(pm.Phase)] = archonPhaseAgent{
			Mode:        "subagent",
			Hidden:      true,
			Model:       pm.Model,
			Variant:     pm.Effort,
			Description: "Archon SDD " + pm.Phase + " phase",
			Prompt:      "{file:./AGENTS.md}",
		}
	}
	doc["agent"] = agents

	// Provider block: coalesce every resolved ref carrying a BaseURL (phases +
	// leader, leader traversed last) into doc["provider"]. Refs without a
	// BaseURL never reach buildProviderBlock, so a config with no local
	// endpoints leaves doc["provider"] completely untouched.
	providerRefs := make([]config.PhaseModel, 0, len(phases)+1)
	providerRefs = append(providerRefs, phases...)
	if leaderFull != "" {
		providerRefs = append(providerRefs, config.PhaseModel{
			Phase:    "leader",
			Model:    leaderFull,
			Provider: models.Leader.Provider,
			Effort:   models.Leader.Effort,
			BaseURL:  models.Leader.BaseURL,
		})
	}
	if block := buildProviderBlock(providerRefs, w); block != nil {
		providers, ok := doc["provider"].(map[string]any)
		if !ok {
			providers = map[string]any{}
		}
		for id, entry := range block {
			providers[id] = entry
		}
		doc["provider"] = providers
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal opencode.json: %w", err)
	}
	data = append(data, '\n')

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return "", fmt.Errorf("write temp opencode.json: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return "", fmt.Errorf("rename opencode.json: %w", err)
	}

	return path, nil
}

// openaiCompatibleNPM is the fixed npm package every local (BaseURL-bearing)
// OpenCode V1 provider block uses. OpenCode V2 uses a different "aisdk:"
// prefix and a "settings.baseURL" shape; this package targets V1 only per the
// project's opencode.json $schema (see design.md "Open Questions" — V2
// divergence is a documented, accepted risk, not abstracted away here).
const openaiCompatibleNPM = "@ai-sdk/openai-compatible"

// buildProviderBlock coalesces every resolved ref carrying a non-empty
// BaseURL into an OpenCode V1 "provider" block, one entry per distinct
// Provider id. Refs are read in the order given (PhaseOrder, leader last, per
// the caller); for each Provider id the FIRST ref's BaseURL wins, and a later
// ref for the same id with a DIFFERENT BaseURL emits a warning to w naming
// the id and the BaseURL that was kept. Every ref's bare model id (Provider
// stripped from FullID) is added to that provider's "models" map — this is
// the union across all refs that share the id.
//
// Returns nil when no ref carries a BaseURL, so the caller can skip touching
// doc["provider"] entirely and existing output stays byte-identical.
//
// Determinism (REQ-5): every level built here is map[string]any, so
// json.MarshalIndent sorts provider ids and model ids lexicographically for
// free — no manual sort is needed.
func buildProviderBlock(refs []config.PhaseModel, w io.Writer) map[string]any {
	type providerEntry struct {
		baseURL string
		models  map[string]any
	}

	entries := make(map[string]*providerEntry)
	for _, ref := range refs {
		if ref.BaseURL == "" {
			continue
		}
		// A ref with an empty Provider id cannot be routed by OpenCode — skip it.
		// ValidateBaseURL already warns about this at config-set time; silently
		// dropping the ref here is consistent with the warn-never-fail contract.
		if ref.Provider == "" {
			continue
		}
		e, ok := entries[ref.Provider]
		if !ok {
			e = &providerEntry{baseURL: ref.BaseURL, models: map[string]any{}}
			entries[ref.Provider] = e
		} else if e.baseURL != ref.BaseURL {
			fmt.Fprintf(w, "warning: provider %q declared with conflicting baseURLs — using first occurrence %q\n", ref.Provider, e.baseURL)
		}
		model := bareModelID(ref)
		e.models[model] = map[string]any{"name": model}
	}

	if len(entries) == 0 {
		return nil
	}

	block := make(map[string]any, len(entries))
	for id, e := range entries {
		block[id] = map[string]any{
			"npm":     openaiCompatibleNPM,
			"options": map[string]any{"baseURL": e.baseURL},
			"models":  e.models,
		}
	}
	return block
}

// bareModelID strips the "<provider>/" prefix FullID() adds so the provider
// block's "models" map key is the bare model id (e.g. "llama3"), matching
// what agent.archon-<phase>.model's own suffix names. A ref whose Model
// already contained "/" before FullID() joining (FullID returns it as-is) or
// whose Provider is empty is returned unchanged.
func bareModelID(pm config.PhaseModel) string {
	if pm.Provider == "" {
		return pm.Model
	}
	prefix := pm.Provider + "/"
	if s, ok := strings.CutPrefix(pm.Model, prefix); ok {
		return s
	}
	return pm.Model
}
