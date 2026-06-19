package initcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// leaderAgentName is the key under "agent" in opencode.json that holds the
// Archon SDD orchestration leader.
const leaderAgentName = "archon-leader"

// archonLeaderAgent is the fixed shape written under agent.archon-leader.
// Struct field declaration order is the deterministic JSON output order.
type archonLeaderAgent struct {
	Mode        string `json:"mode"`        // always "primary"
	Description string `json:"description"` // fixed cosmetic label
	Model       string `json:"model"`       // verbatim models.leader
	Prompt      string `json:"prompt"`      // "{file:./AGENTS.md}"
}

// MergeOpencodeAgent is the exported integration seam for callers outside this
// package (notably the TUI save path). It delegates to mergeOpencodeAgent so
// init and the TUI share a single writer implementation and produce
// byte-identical opencode.json output.
func MergeOpencodeAgent(projectDir, leader string) (written string, err error) {
	return mergeOpencodeAgent(projectDir, leader)
}

// mergeOpencodeAgent additively merges agent.archon-leader into
// <projectDir>/opencode.json. It is a no-op (returns "", nil) when leader == "".
//
// Existing top-level keys and existing agents are preserved; only
// agent.archon-leader is set. It never writes a default_agent key. The whole
// document is marshaled with json.MarshalIndent (sorted map keys) plus a
// trailing newline, then written atomically via a temp file + os.Rename — the
// same pattern as config.Save and writeTemplate — so re-runs yield byte-
// identical output. It returns the written path ("" when nothing was written)
// so the caller can register it for rollback.
func mergeOpencodeAgent(projectDir, leader string) (written string, err error) {
	if leader == "" {
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
	agents[leaderAgentName] = archonLeaderAgent{
		Mode:        "primary",
		Description: "Archon SDD orchestration leader",
		Model:       leader,
		Prompt:      "{file:./AGENTS.md}",
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
