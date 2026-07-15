# Tasks: Phase-Model Hard Gate for the Claude Harness

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~270–340 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: stacked-to-main
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Full change (writer + wiring + templates + tests) | PR 1 | Single PR to master; ~270–340 lines, under 400-line budget |

---

## Phase 1: Foundation

- [x] 1.1 Read `internal/initcmd/opencode_mode.go` and `internal/initcmd/templates.go` to baseline the patterns before touching any file.
- [x] 1.2 Verify `config.ResolvePhaseModels`, `config.PhaseOrder`, and `config.ModelConfig` are importable from `internal/initcmd` (confirm existing usages in `init.go`).

## Phase 2: Core Writer

- [x] 2.1 Create `internal/initcmd/claude_mode.go`: package `initcmd`; unexported `writeClaudeAgents(projectDir string, models config.ModelConfig) ([]string, error)` iterating `config.ResolvePhaseModels(models)` — early-return `nil, nil` when `len(phases) == 0`; call `os.MkdirAll(.claude/agents, 0o755)` inside the loop only; atomic temp+rename per file.
- [x] 2.2 In the same file, add `renderClaudeAgent(pm config.PhaseModel) []byte` producing fixed-order YAML frontmatter (`name`, `description`, `model: <FullID>`) + body `"You are the Archon SDD <phase> executor. Follow \`skills/sdd-<phase>/SKILL.md\`…"` + trailing newline.
- [x] 2.3 Export `WriteClaudeAgents(projectDir string, models config.ModelConfig) ([]string, error)` delegating to `writeClaudeAgents` (TUI seam, mirrors `MergeOpencodeAgent`).

## Phase 3: Wiring

- [x] 3.1 Edit `internal/initcmd/init.go`: after the opencode branch (~line 101), add `if agentName == "claude"` gate calling `writeClaudeAgents`; append returned paths to `rollback.CreatedPaths` before `WriteManifest`.
- [x] 3.2 Edit `internal/tui/model.go` `saveConfig`: add `if cfg.Agent == "claude"` branch calling `initcmd.WriteClaudeAgents(projectDir, cfg.Models)` (no rollback, matching opencode's TUI branch).

## Phase 4: Template Split

- [x] 4.1 Edit `internal/initcmd/templates.go`: extract `orchestratorTrailerHead` (shared content before the Phase Models block); replace single `## Phase Models` with `phaseModelsClaude` ("the `archon-<phase>` subagents bind each phase's model — a hard gate, not advisory") and `phaseModelsOpencode` ("the binding lives in `opencode.json`").
- [x] 4.2 In the same file, split the delegation rule (Rule 2) into per-harness variants: claude variant names `archon-<phase>` as the delegation target and adds "do not pass a per-call model parameter"; opencode variant names `archon-<phase>` without that constraint. Neither variant mentions `CLAUDE_CODE_SUBAGENT_MODEL` or the word "advisory".

## Phase 5: Tests

- [x] 5.1 Create `internal/initcmd/claude_mode_test.go` (mirrors `opencode_mode_test.go`): table-driven test for `writeClaudeAgents` covering scenarios "Init writes an agent file per resolvable phase" and "A phase with no resolvable model is omitted".
- [x] 5.2 Add test for scenario "Frontmatter model matches the resolved FullID": parse frontmatter `model:` field from written file; assert equals `pm.Model`.
- [x] 5.3 Add test for scenario "Phase falls back to the default model": set `models.default`, leave all `phases` empty; assert `archon-<phase>.md` frontmatter `model` equals the default FullID.
- [x] 5.4 Add test for scenario "Body points the executor at the phase skill": assert written body contains `"skills/sdd-<phase>/SKILL.md"` and is non-empty after frontmatter.
- [x] 5.5 Add tests for scenarios "Nothing resolvable writes nothing" and "Non-claude agent writes no claude agent files": stat `.claude/agents`; assert `os.IsNotExist`.
- [x] 5.6 Add test for scenario "Re-run is byte-identical and preserves user files": seed a user file + run writer twice; assert byte equality per archon file and user file unchanged.
- [x] 5.7 Add integration test for scenario "Undo removes the generated agent files": call `Run(agent=claude)`; call rollback; assert all `archon-<phase>.md` paths absent; load manifest via `LoadManifest` to verify registration.
- [x] 5.8 Edit `internal/initcmd/templates_test.go`: assert scenario "CLAUDE.md names subagents as the hard gate" (contains `archon-<phase>`, absent "advisory"); assert scenario "AGENTS.md points at opencode.json" (contains `opencode.json`); assert scenario "CLAUDE.md routes delegation to the named subagent" (contains `archon-<phase>` + no per-call model param instruction); assert scenario "AGENTS.md routes delegation to the named subagent".
