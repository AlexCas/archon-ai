package initcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/archon-ai/archon/internal/config"
)

// WriteClaudeAgents is the exported integration seam for callers outside this
// package (notably the TUI save path). It delegates to writeClaudeAgents so
// init and the TUI share a single writer implementation and produce
// byte-identical .claude/agents output.
func WriteClaudeAgents(projectDir string, models config.ModelConfig) ([]string, error) {
	return writeClaudeAgents(projectDir, models)
}

// writeClaudeAgents writes one .claude/agents/archon-<phase>.md file per
// resolvable SDD phase into <projectDir>/.claude/agents/. It is a no-op
// (returns nil, nil) when ResolvePhaseModels returns no phases (no-op must not
// create the directory). The written paths are returned so the caller can
// register them for rollback. The write is atomic (temp file + os.Rename) and
// idempotent — re-running with the same config produces byte-identical files.
// Only archon-<phase>.md files are written; user-defined files under
// .claude/agents/ are never modified or deleted.
func writeClaudeAgents(projectDir string, models config.ModelConfig) ([]string, error) {
	phases := config.ResolvePhaseModels(models)
	if len(phases) == 0 {
		return nil, nil
	}

	dir := filepath.Join(projectDir, ".claude", "agents")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create claude agents dir: %w", err)
	}

	var written []string
	for _, pm := range phases {
		path := filepath.Join(dir, "archon-"+pm.Phase+".md")
		content := renderClaudeAgent(pm)

		tmp := path + ".tmp"
		if err := os.WriteFile(tmp, content, 0o644); err != nil {
			return nil, fmt.Errorf("write temp agent file for %s: %w", pm.Phase, err)
		}
		if err := os.Rename(tmp, path); err != nil {
			return nil, fmt.Errorf("rename agent file for %s: %w", pm.Phase, err)
		}
		written = append(written, path)
	}

	return written, nil
}

// renderClaudeAgent produces the fixed-order YAML frontmatter + functional body
// for a Claude Code subagent file. Field order is fixed (name, description,
// model) so re-runs with the same PhaseModel yield byte-identical output. The
// file ends with a single trailing newline.
func renderClaudeAgent(pm config.PhaseModel) []byte {
	content := "---\n"
	content += "name: archon-" + pm.Phase + "\n"
	content += "description: Archon SDD " + pm.Phase + " phase executor\n"
	content += "model: " + claudeFrontmatterModel(pm.Model) + "\n"
	content += "---\n"
	content += "\n"
	content += "You are the Archon SDD " + pm.Phase + " executor. Follow `skills/sdd-" + pm.Phase + "/SKILL.md`\n"
	content += "for this phase. Do NOT delegate; execute the phase yourself.\n"
	return []byte(content)
}

// claudeFrontmatterModel converts a resolved FullID into the value Claude Code
// accepts in a subagent's `model:` frontmatter. ResolvePhaseModels emits the
// opencode-style "<provider>/<model>" FullID, but Claude Code's model field
// expects a bare model id (e.g. "claude-opus-4-8") or alias — a provider prefix
// makes it reject the model. We therefore drop everything up to and including
// the last "/". A bare alias (no "/") is returned unchanged.
func claudeFrontmatterModel(fullID string) string {
	if i := strings.LastIndex(fullID, "/"); i >= 0 {
		return fullID[i+1:]
	}
	return fullID
}
