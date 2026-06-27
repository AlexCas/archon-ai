# Verification Report

**Change**: phase-model-hard-gate
**Version**: N/A (delta spec `claude-phase-subagents`)
**Mode**: Standard (no `strict_tdd`; Go CLI, non-web → no Playwright)

## Verification Report

### Completeness
| Metric | Value |
|--------|-------|
| Tasks total | 17 |
| Tasks complete | 17 |
| Tasks incomplete | 0 |

All 17 tasks across Phases 1–5 are marked `[x]`. No unchecked implementation tasks.

### Build & Tests Execution

**Build**: ✅ Passed
```text
$ go build ./...
EXIT=0
```

**Vet**: ✅ Passed
```text
$ go vet ./internal/initcmd/... ./internal/tui/...
EXIT=0
```

**Tests**: ✅ All passed / ❌ 0 failed / ⚠️ 0 skipped
```text
$ go test ./internal/initcmd/... ./internal/tui/... -count=1 -v
ok  github.com/archon-ai/archon/internal/initcmd  0.054s
ok  github.com/archon-ai/archon/internal/tui      2.426s

Claude-specific tests (all PASS):
--- PASS: TestWriteClaudeAgents_WritesOneFilePerResolvablePhase
--- PASS: TestWriteClaudeAgents_OmitsPhaseWithNoModel
--- PASS: TestWriteClaudeAgents_FrontmatterModelMatchesResolvedFullID
--- PASS: TestWriteClaudeAgents_PhaseFallsBackToDefault
--- PASS: TestWriteClaudeAgents_BodyPointsAtPhaseSkill
--- PASS: TestWriteClaudeAgents_NothingResolvableWritesNothing
--- PASS: TestRun_NonClaudeWritesNoClaudeAgentFiles
--- PASS: TestWriteClaudeAgents_ReRunByteIdenticalAndPreservesUserFiles
--- PASS: TestRun_ClaudeRegistersAgentPathsForRollback
--- PASS: TestTemplates_ClaudePhaseModelsIsHardGate
--- PASS: TestTemplates_AgentsPhaseModelsPointsAtOpencodeJSON
--- PASS: TestTemplates_ClaudeDelegationRuleNamesSubagentAndNoPerCallModel
--- PASS: TestTemplates_AgentsDelegationRuleNamesSubagent
```

Full repo suite (`go test ./... -count=1`) is green across all 11 packages — no regression.

**Coverage**: 53.7% of initcmd statements (package-level, `-run Claude`) → ➖ no configured threshold (`coverage_threshold: 0`). Signal only.

### Spec Compliance Matrix
| Requirement | Scenario | Test | Result |
|-------------|----------|------|--------|
| Per-phase emission | Init writes an agent file per resolvable phase | `claude_mode_test.go > TestWriteClaudeAgents_WritesOneFilePerResolvablePhase` (asserts file per resolved phase + no `archon-judge.md`) | ✅ COMPLIANT |
| Per-phase emission | A phase with no resolvable model is omitted | `claude_mode_test.go > TestWriteClaudeAgents_OmitsPhaseWithNoModel` | ✅ COMPLIANT |
| Frontmatter FullID | Frontmatter model matches the resolved FullID | `claude_mode_test.go > TestWriteClaudeAgents_FrontmatterModelMatchesResolvedFullID` | ✅ COMPLIANT |
| Frontmatter FullID | Phase falls back to the default model | `claude_mode_test.go > TestWriteClaudeAgents_PhaseFallsBackToDefault` | ✅ COMPLIANT |
| Functional body | Body points the executor at the phase skill | `claude_mode_test.go > TestWriteClaudeAgents_BodyPointsAtPhaseSkill` (substring + non-empty after frontmatter) | ✅ COMPLIANT |
| No-op condition | Nothing resolvable writes nothing | `claude_mode_test.go > TestWriteClaudeAgents_NothingResolvableWritesNothing` (`.claude/agents` absent) | ✅ COMPLIANT |
| No-op condition | Non-claude agent writes no claude agent files | `claude_mode_test.go > TestRun_NonClaudeWritesNoClaudeAgentFiles` (Run with `--agent opencode`) | ✅ COMPLIANT |
| Determinism/preservation | Re-run is byte-identical and preserves user files | `claude_mode_test.go > TestWriteClaudeAgents_ReRunByteIdenticalAndPreservesUserFiles` | ✅ COMPLIANT |
| Rollback registration | Undo removes the generated agent files | `claude_mode_test.go > TestRun_ClaudeRegistersAgentPathsForRollback` (manifest registration) + `config/rollback_test.go > TestRollbackManifest_Cleanup` (removal of registered paths) | ✅ COMPLIANT (transitive — see WARNING) |
| Doc names binding | CLAUDE.md names subagents as the hard gate | `templates_test.go > TestTemplates_ClaudePhaseModelsIsHardGate` (`archon-<phase>` present, "advisory" absent) | ✅ COMPLIANT |
| Doc names binding | AGENTS.md points at opencode.json | `templates_test.go > TestTemplates_AgentsPhaseModelsPointsAtOpencodeJSON` | ✅ COMPLIANT |
| Delegation routing | CLAUDE.md routes delegation to the named subagent | `templates_test.go > TestTemplates_ClaudeDelegationRuleNamesSubagentAndNoPerCallModel` (`archon-<phase>` + "per-call model parameter") | ✅ COMPLIANT |
| Delegation routing | AGENTS.md routes delegation to the named subagent | `templates_test.go > TestTemplates_AgentsDelegationRuleNamesSubagent` | ✅ COMPLIANT |

**Compliance summary**: 13/13 scenarios compliant.

### Correctness (Static Evidence)
| Requirement | Status | Notes |
|------------|--------|-------|
| Per-phase agent file emission | ✅ Implemented | `writeClaudeAgents` iterates `config.ResolvePhaseModels`, writes one `archon-<phase>.md` per resolved phase. `judge` never in `PhaseOrder` ⇒ no `archon-judge.md` (asserted). |
| No `archon-judge` | ✅ Implemented | Verified by test stat-asserting `archon-judge.md` absent. |
| Frontmatter FullID | ✅ Implemented | `renderClaudeAgent` emits `model: <pm.Model>` (FullID) in fixed order `name`,`description`,`model`. |
| Default fallback | ✅ Implemented | `ResolvePhaseModels` chains phase→default; fallback test confirms `archon-tasks.md` carries default FullID. |
| Functional body referencing phase skill | ✅ Implemented | Body: "You are the Archon SDD `<phase>` executor. Follow `skills/sdd-<phase>/SKILL.md`… Do NOT delegate; execute the phase yourself." Non-empty, references skill path. |
| No-op + non-claude | ✅ Implemented | Early `return nil, nil` on zero phases (dir not created); init gate `agentName == "claude"` and TUI gate `cfg.Agent == "claude"` confine writes to claude. |
| Determinism / non-clobber | ✅ Implemented | Fixed-order frontmatter + single trailing newline; atomic `WriteFile(tmp)`+`Rename`; only `archon-<phase>.md` written, user files untouched (test seeds `my-custom-agent.md`, verifies preserved). |
| Rollback registration | ✅ Implemented | init appends written paths to `rollback.CreatedPaths` BEFORE `WriteManifest`; `Cleanup` removes them. TUI branch intentionally omits rollback (mirrors opencode). |
| Doc wording — CLAUDE.md Phase Models | ✅ Implemented | Rendered block names `archon-<phase>` subagents as binding hard gate; word "advisory" absent (rendered output verified). |
| Doc wording — AGENTS.md Phase Models | ✅ Implemented | Rendered block: "binding lives in `opencode.json` under `agent.archon-<phase>.model`". |
| Delegation rule rewrite — CLAUDE.md | ✅ Implemented | Rule 2: "Delegate each phase to the `archon-<phase>` subagent; do not pass a per-call model parameter — the subagent's frontmatter model is the gate." |
| Delegation rule rewrite — AGENTS.md | ✅ Implemented | Rule 2: "Delegate each phase to the `archon-<phase>` subagent" (no per-call constraint, per spec). |
| No `CLAUDE_CODE_SUBAGENT_MODEL` in docs | ✅ Implemented | Grep of generated templates: env var absent. The only file occurrences are a Go source comment ("not an advisory preference") and test assertions — neither is emitted to generated docs. |

### Coherence (Design — 7 decisions)
| Decision | Followed? | Notes |
|----------|-----------|-------|
| 1. One `.md` per phase, atomic temp+rename | ✅ Yes | Per-file `WriteFile(path+".tmp")` then `Rename`. |
| 2. Writer signature `([]string, error)` feeding rollback | ✅ Yes | `writeClaudeAgents(projectDir, models) ([]string, error)`; slice appended to `CreatedPaths`. |
| 3. Frontmatter `name`/`description`/`model`, no effort/variant | ✅ Yes | Exactly those three fields, fixed order; no variant key. |
| 4. `MkdirAll` only inside write loop (skip on no-op) | ✅ Yes | `MkdirAll` after the `len(phases)==0` early return; no-op test confirms dir absent. |
| 5. Non-clobber — write only `archon-<phase>.md` | ✅ Yes | No list/delete of other files; user-file preservation test passes. |
| 6. Template split — shared head + per-harness block | ✅ Yes (restructured, faithful) | See note below. |
| 7. Delegation routing rewrite naming `archon-<phase>` | ✅ Yes | Both docs' Rule 2 rewritten; CLAUDE.md adds the per-call-model constraint. |

**Decision 6 restructure note**: The apply agent split the single `orchestratorTrailer` into `orchestratorRulesClaude`/`orchestratorRulesOpencode` + `orchestratorTrailerHead` (Configuration) + `phaseModelsClaude`/`phaseModelsOpencode` + `orchestratorStateManagement`, assembled as `orchestratorSections + orchestratorRules{X} + orchestratorTrailerHead + phaseModels{X} + orchestratorStateManagement`. This is faithful to the design intent (shared sections preserved; only Rules and Phase Models diverge per harness) and goes one step further by also splitting Rules (required to host the per-harness Rule 2). Rendered output confirms both docs carry identical `orchestratorSections`, Configuration, and State Management — shared sections did NOT drift. Not a regression.

### Issues Found

**CRITICAL**: None.

**WARNING**:
- *Undo scenario tested transitively, not end-to-end.* Task 5.7 / scenario "Undo removes the generated agent files" is verified by (a) `TestRun_ClaudeRegistersAgentPathsForRollback` proving the paths land in `manifest.CreatedPaths`, and (b) the pre-existing `TestRollbackManifest_Cleanup` proving `Cleanup` removes registered `CreatedPaths` via `os.RemoveAll`. There is no single test that runs `Run(agent=claude)` then `Cleanup()` and stat-asserts the `archon-<phase>.md` files are gone. The two facts compose correctly and the mechanism is sound, but a direct end-to-end undo assertion would make the scenario self-evidently green without relying on a generic rollback test. Non-blocking.

**SUGGESTION**:
- *Missing blank line before `## Rules`.* `orchestratorSections` ends with `…not the tool.\n` and `orchestratorRules{Claude,Opencode}` begins immediately with `## Rules`, so the rendered docs show `…not the tool.\n## Rules` with no blank separator line. Every other section is separated by a blank line. Cosmetic markdown spacing only — does not affect any spec assertion or harness behavior. Consider prefixing the Rules constants with a leading `\n`.
- *Rollback `HomeDir` left empty in `buildRollbackManifest`* (pre-existing, out of scope): the manifest path resolves relative to cwd rather than an explicit home. The claude test passes under the controlled test environment and this mirrors the existing opencode rollback pattern; flagged only for awareness, not introduced by this change.

### Verdict
**PASS WITH WARNINGS**

All 17 tasks complete, build/vet/full test suite green, all 13 Gherkin scenarios mapped to passing covering tests, all 8 spec requirements and 7 design decisions satisfied. Generated docs are correct (no "advisory", no `CLAUDE_CODE_SUBAGENT_MODEL`, `archon-<phase>` named in both Phase Models and delegation rules). The single WARNING is a test-strength observation on the undo scenario (covered transitively, not end-to-end), not a correctness gap; the SUGGESTIONS are cosmetic/pre-existing.
