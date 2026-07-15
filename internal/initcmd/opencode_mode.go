package initcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
func MergeOpencodeAgent(projectDir string, models config.ModelConfig) (written string, err error) {
	return mergeOpencodeAgent(projectDir, models)
}

// mergeOpencodeAgent additively merges agent.archon-leader and one
// agent.archon-<phase> subagent per resolvable SDD phase into
// <projectDir>/opencode.json. It is a no-op (returns "", nil) when the leader
// FullID is empty and ResolvePhaseModels returns no phases.
//
// Existing top-level keys and existing agents are preserved; only
// agent.archon-leader and agent.archon-<phase> entries are set. It never
// writes a default_agent key. The whole document is marshaled with
// json.MarshalIndent (sorted map keys) plus a trailing newline, then written
// atomically via a temp file + os.Rename — the same pattern as config.Save and
// writeTemplate — so re-runs yield byte-identical output. It returns the
// written path ("" when nothing was written) so the caller can register it for
// rollback.
func mergeOpencodeAgent(projectDir string, models config.ModelConfig) (written string, err error) {
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
