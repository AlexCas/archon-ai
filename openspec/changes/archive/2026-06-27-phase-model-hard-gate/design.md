# Design: Phase-Model Hard Gate for the Claude Harness

## Technical Approach

Mirror the proven `opencode_mode.go` writer for claude. A new
`internal/initcmd/claude_mode.go` exposes `writeClaudeAgents(projectDir, models)
([]string, error)` (exported `WriteClaudeAgents` as the TUI seam). It iterates
`config.ResolvePhaseModels(models)` and writes one
`.claude/agents/archon-<phase>.md` file per resolvable phase, each with fixed
YAML frontmatter (`model: <FullID>`) and a functional body referencing
`skills/sdd-<phase>/SKILL.md`. Init and TUI gate it on `agent == "claude"`,
registering every written path for rollback. The shared `orchestratorTrailer` is
split so CLAUDE.md and AGENTS.md carry different "Phase Models" wording while all
other sections stay shared. Satisfies all 7 requirements in
`specs/claude-phase-subagents/spec.md`.

## Architecture Decisions

| # | Decision | Choice | Rationale |
|---|----------|--------|-----------|
| 1 | File layout | One `.md` file per phase (vs opencode's single JSON) | Claude Code resolves subagents from individual `.claude/agents/*.md` files; there is no merge target. Per-file atomic temp+rename gives the same determinism without a parse/merge step. |
| 2 | Writer signature | `writeClaudeAgents(projectDir string, models config.ModelConfig) ([]string, error)` returning all written paths | Mirrors opencode's `(written string, err)` but plural because N files are produced; the slice feeds `rollback.CreatedPaths` directly. |
| 3 | Frontmatter fields | `name`, `description`, `model` in fixed order; no effort/variant key | Claude Code subagent frontmatter has no variant field (unlike opencode's `variant`). Effort is intentionally dropped; FullID alone is the binding. |
| 4 | Dir creation | `os.MkdirAll(.claude/agents, 0o755)` only inside the write loop (skipped on no-op) | Spec: no-op must create nothing. Avoids an empty `agents/` dir when no models resolve. |
| 5 | Non-clobber | Write only `archon-<phase>.md`; never list/delete other files | Spec requirement "Deterministic output and preservation". User files in `.claude/agents/` are untouched. |
| 6 | Template split | Split trailer into shared head + per-agent Phase Models block | Keeps `orchestratorSections` and the rest of the trailer shared (drift risk), diverging only the block that must differ per harness. |
| 7 | Delegation routing | Rewrite the leader's delegation rule in both docs to name `archon-<phase>` as the per-phase delegation target, with "do not pass a per-call model param" | The agent-definition gate only fires when the leader invokes the named agent. Rule 2 ("delegate to sdd-* sub-agent") names the skill, not the model-bound agent — closing this gap makes the gate reachable end-to-end. |

## Data Flow

    init.go (agent=="claude")  ┐
    tui/model.go saveConfig    ┘──→ WriteClaudeAgents(projectDir, cfg.Models)
                                       │
              config.ResolvePhaseModels(models)  →  []PhaseModel
                                       │ per phase
              .claude/agents/archon-<phase>.md  (temp + rename)
                                       │
              []paths ──→ rollback.CreatedPaths ──→ WriteManifest

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/initcmd/claude_mode.go` | Create | `writeClaudeAgents` + exported `WriteClaudeAgents`; per-file deterministic writer |
| `internal/initcmd/templates.go` | Modify | Split `orchestratorTrailer`; per-harness Phase Models block + delegation rule (name `archon-<phase>`) |
| `internal/initcmd/init.go` | Modify | `agentName == "claude"` gate after template write; append paths to rollback |
| `internal/tui/model.go` | Modify | `cfg.Agent == "claude"` branch in `saveConfig` calling `WriteClaudeAgents` |
| `internal/initcmd/claude_mode_test.go` | Create | Mirror of `opencode_mode_test.go` |
| `internal/initcmd/templates_test.go` | Modify | Assert per-harness Phase Models wording |

## Interfaces / Contracts

Frontmatter + body template (fixed field order, single trailing newline,
byte-identical on re-run):

```markdown
---
name: archon-<phase>
description: Archon SDD <phase> phase executor
model: <FullID>
---

You are the Archon SDD <phase> executor. Follow `skills/sdd-<phase>/SKILL.md`
for this phase. Do NOT delegate; execute the phase yourself.
```

Writer skeleton (per-file atomic write; no-op guard mirrors opencode):

```go
func writeClaudeAgents(projectDir string, models config.ModelConfig) ([]string, error) {
    phases := config.ResolvePhaseModels(models)
    if len(phases) == 0 {
        return nil, nil
    }
    dir := filepath.Join(projectDir, ".claude", "agents")
    if err := os.MkdirAll(dir, 0o755); err != nil { ... }
    var written []string
    for _, pm := range phases {
        path := filepath.Join(dir, "archon-"+pm.Phase+".md")
        content := renderClaudeAgent(pm) // fixed-order frontmatter + body + "\n"
        // os.WriteFile(path+".tmp"); os.Rename(tmp, path)
        written = append(written, path)
    }
    return written, nil
}
```

Init wiring (parallel to opencode at init.go:101):

```go
if agentName == "claude" {
    paths, err := writeClaudeAgents(opts.ProjectDir, cfg.Models)
    if err != nil { return nil, fmt.Errorf("write claude agents: %w", err) }
    rollback.CreatedPaths = append(rollback.CreatedPaths, paths...)
}
```

TUI: add `cfg.Agent == "claude"` branch in `saveConfig` calling
`initcmd.WriteClaudeAgents` (no rollback there, matching opencode's TUI branch).

Template split: replace the single `## Phase Models` block in
`orchestratorTrailer` with `orchestratorTrailerHead` (shared) +
`phaseModelsClaude` ("the `archon-<phase>` subagents bind each phase's model — a
hard gate") and `phaseModelsOpencode` ("the binding lives in `opencode.json`").
Neither uses the word "advisory" nor mentions `CLAUDE_CODE_SUBAGENT_MODEL`.

Delegation rule rewrite (same per-harness split): the leader's delegation rule
(currently Rule 2, "Delegate each phase to sdd-* sub-agent") MUST name the
`archon-<phase>` subagent as the per-phase delegation target. The claude variant
adds "do not pass a per-call model parameter — the subagent's frontmatter model is
the gate". The `archon-<phase>` body in turn follows `skills/sdd-<phase>/SKILL.md`,
so the skill chain is preserved. This closes the loop: the agent-definition gate
only fires when the leader invokes the named agent.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|--------------|----------|
| Unit | per-phase emission; no `archon-judge.md`; omitted phase | `writeClaudeAgents` + read dir |
| Unit | frontmatter `model` == resolved FullID; default fallback | parse frontmatter |
| Unit | body references `skills/sdd-<phase>/SKILL.md`, non-empty | substring check |
| Unit | no-op (empty config) and non-claude write nothing | stat `.claude/agents` absent |
| Unit | byte-identical re-run; unrelated user file preserved | seed file + double run |
| Integration | `Run(agent=claude)` registers all paths for rollback | `LoadManifest` |
| Template | CLAUDE.md names subagents (no "advisory"); AGENTS.md names `opencode.json` | render + substring |
| Template | both docs' delegation rule names `archon-<phase>`; CLAUDE.md says no per-call model param | render + substring |

## Migration / Rollout

No migration required. Config types reused as-is; template split is a pure source
change. Rollback removes all generated agent files via the manifest.

## Open Questions

- [ ] None blocking. Effort/variant is deliberately excluded from claude
  frontmatter (no platform field); the leader inherits it only via the body. If a
  future Claude Code release adds a variant field, extend the frontmatter then.
