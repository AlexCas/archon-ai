package initcmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/archon-ai/archon/internal/config"
)

// WriteClaudeAgents is the exported integration seam for callers outside this
// package (notably the TUI save path). It delegates to writeClaudeAgents so
// init and the TUI share a single writer implementation and produce
// byte-identical .claude/agents output.
//
// Caller audit (PR-A A1, local-model-provider): grepping every exported
// symbol in this package across the repo shows exactly one external caller —
// internal/tui/model.go (line ~381, passes io.Discard). init.go is the
// same-package caller of the unexported writeClaudeAgents and passes
// os.Stderr. No third caller exists.
func WriteClaudeAgents(projectDir string, models config.ModelConfig, w io.Writer) ([]string, error) {
	return writeClaudeAgents(projectDir, models, w)
}

// writeClaudeAgents writes one .claude/agents/archon-<phase>.md file per
// resolvable SDD phase into <projectDir>/.claude/agents/. It is a no-op
// (returns nil, nil) when ResolvePhaseModels returns no phases (no-op must not
// create the directory). The written paths are returned so the caller can
// register them for rollback. The write is atomic (temp file + os.Rename) and
// idempotent — re-running with the same config produces byte-identical files.
// Only archon-<phase>.md files are written; user-defined files under
// .claude/agents/ are never modified or deleted.
//
// REQ-6 (PR-B): the Claude path cannot honor a resolved ref's BaseURL — Claude
// Code subagents do not support custom endpoints. For every phase whose
// resolved PhaseModel.BaseURL != "", writeClaudeAgents emits a visible warning
// to w before writing the agent file, then writes the file anyway with the
// bare model id (warn-and-skip: never a hard failure). Callers on the CLI path
// (init.go) pass os.Stderr so the warning is user-visible; the TUI save path
// intentionally passes io.Discard because the warning has no display target in
// the terminal UI and errors are propagated via the error return.
func writeClaudeAgents(projectDir string, models config.ModelConfig, w io.Writer) ([]string, error) {
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
		if pm.BaseURL != "" {
			fmt.Fprintf(w, "warning: phase %q has base_url set but agent is \"claude\" — local endpoint ignored; claude agents do not support custom baseURLs\n", pm.Phase)
		}
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
	if pm.Phase == "judge" {
		content += "You are the Archon SDD judge executor. There is no sdd-judge skill: your job is the\n"
		content += "dual adversarial review. Run the `judgment-day` skill against the current change\n"
		content += "(all files modified by the change), then report its verdict (APPROVED or ESCALATED,\n"
		content += "with confirmed/suspect issues) back to `harness-judge`. Do NOT apply fixes or\n"
		content += "re-verify yourself — harness-judge owns the re-apply loop and the gates.\n"
		return []byte(content)
	}
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
