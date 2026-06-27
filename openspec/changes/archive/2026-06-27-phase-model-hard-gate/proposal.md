# Proposal: Phase-Model Hard Gate for the Claude Harness

## Intent

Per-phase model selection is not a hard gate on the **claude** harness. The shared "Phase Models" block (`templates.go:163-171`) calls it "advisory, not a hard gate". opencode has a real gate (`opencode_mode.go` writes `archon-<phase>` subagents into `opencode.json` with fixed models); claude has **no** per-phase writer, so the leader runs every phase on its inherited model. This is Slice 2 ("Writers") of `structured-model-resolution`, extended to claude.

## Scope

### In Scope
- New writer `claude_mode.go`: emit `.claude/agents/archon-<phase>.md` per phase with frontmatter `model: <FullID>` + a functional body.
- Hook into `init.go` and `tui/model.go` saveConfig (gated on `claude`); register written paths for rollback.
- Split `orchestratorTrailer`: `CLAUDE.md`'s block names `archon-<phase>` subagents; `AGENTS.md`'s points at `opencode.json`.
- Rewrite the delegation rule in BOTH docs so the leader delegates each phase to the named `archon-<phase>` subagent (no per-call model param) — closes the hard-gate loop end-to-end.

### Out of Scope
- Behavioral change to the opencode writer (`mergeOpencodeAgent`).
- `judge` phase (not in `PhaseOrder`).
- New config types — `ResolvePhaseModels`/`FullID` reused as-is.

## Capabilities

### New Capabilities
- `claude-phase-subagents`: init + TUI save write one `.claude/agents/archon-<phase>.md` per resolvable phase, binding its FullID via frontmatter `model:` — mirroring `opencode-phase-writers`.

### Modified Capabilities
- None at the spec level. The "Phase Models" template rewrite ships inside `claude-phase-subagents`.

## Approach

Approach C (exploration): mirror `opencode_mode.go`. `writeClaudeAgents(projectDir, models) ([]string, error)` iterates `config.ResolvePhaseModels(models)`, atomically writes each agent file (no-op on zero phases, idempotent), returns written paths. Export `WriteClaudeAgents` as the TUI seam. Each body follows `skills/sdd-<phase>/SKILL.md`. Keep `orchestratorSections` shared; split only the trailer.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `initcmd/claude_mode.go` | New | Per-phase writer + exported seam |
| `initcmd/templates.go` | Modified | Split trailer; claude vs opencode block |
| `initcmd/init.go` | Modified | `claude` gate + rollback registration |
| `tui/model.go` | Modified | `claude` gate in saveConfig |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Template blocks drift after split | Med | Keep `orchestratorSections` shared; split only trailer |
| Orphan files after `archon undo` | Med | Append written paths to `rollback.CreatedPaths` before `WriteManifest` |
| Non-idempotent re-runs | Low | Atomic temp+rename, as opencode |
| Empty dir on no models | Low | No-op when `len(phases) == 0` |

## Rollback Plan

Written paths are registered in the rollback manifest, so `archon undo` removes the generated agent files. No migration or persisted state; the template split is a pure source revert.

## Dependencies

- `config/model.go` (`ResolvePhaseModels`, `PhaseOrder`, `FullID`) — shipped (PR #47).

## Success Criteria

- [ ] `archon init --agent claude` writes one `archon-<phase>.md` per resolvable phase; none for `judge`.
- [ ] Each frontmatter `model:` equals the resolved FullID; no-op when none resolve; re-runs byte-identical.
- [ ] `archon undo` removes all generated agent files.
- [ ] `CLAUDE.md` names `archon-<phase>` subagents (no "advisory"); `AGENTS.md` points at `opencode.json`; opencode output unchanged.
- [ ] Both docs' delegation rule names `archon-<phase>` as the per-phase target; `CLAUDE.md` says not to pass a per-call model param.

## Proposal question round (CONFIRMED 2026-06-27)

1. **Agent body content** — CONFIRMED: each `.claude/agents/archon-<phase>.md` body instructs the executor to follow `skills/sdd-<phase>/SKILL.md` (functional system prompt, not an empty stub).
2. **Doc-fix scope** — CONFIRMED: fix the "Phase Models" wording for **both** harnesses (shared template) with **no behavioral** change to the opencode writer.
3. **Positioning** — CONFIRMED: ship as Slice 2 ("Writers") of `structured-model-resolution`, extended to claude.

## Size / PR forecast

~250–320 changed lines (writer ~110, gates ~20, template split ~30, tests ~120). Under the **400-line** budget (D1) → **single PR** (C1). `400-line budget risk: Low`.
