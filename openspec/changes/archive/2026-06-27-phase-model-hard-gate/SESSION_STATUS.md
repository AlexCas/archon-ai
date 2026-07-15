# SESSION STATUS

**Active change**: phase-model-hard-gate
**Current phase**: judge
**Status**: completed
**Last updated**: 2026-06-27T08:50:00Z
**Branch**: feat/phase-model-hard-gate (off master)
**Note**: root CLAUDE.md has a PRE-EXISTING (pre-session) modification with the OLD advisory wording — NOT part of this feature; exclude from the feature commit. Optional follow-up: regenerate root CLAUDE.md/AGENTS.md via `archon init` to dogfood the corrected template.

## Preflight Choices
- **A. Rhythm**: A1 — Interactive (show each phase, wait for confirmation)
- **B. Artifacts**: B1 — OpenSpec (files in repo)
- **C. PRs**: C1 — Ask-always (stop if estimate exceeds budget)
- **D. Review budget**: D1 — 400 changed lines
- **E. Playwright**: E1 — Disabled (Go CLI, not web)

## Problem Summary
Per-phase model selection does not work as a hard gate for subagents.
- **opencode**: hard gate exists in `opencode.json` (per-phase `archon-<phase>` subagents with fixed `model`), but the generated "Phase Models" doc text calls it "advisory" and tells the leader to pass `model: <id>` at call time — which opencode ignores. Leader ends up running phases on the leader model.
- **claude**: NO per-phase subagents are generated at all (no `.claude/agents/archon-*.md` writer in `internal/initcmd`). Claude Code DOES support real per-phase model binding (frontmatter `model:` accepts full IDs; Task tool `model` param is enforced), but archon never emits the agent files.
- **Shared root cause**: `internal/initcmd/templates.go:163-167` generates one misleading "Phase Models / not a hard gate" block for both CLAUDE.md and AGENTS.md.

## Intended Fix (scope)
1. Add a claude per-phase subagent writer (mirror of `mergeOpencodeAgent` in `opencode_mode.go`) → emit `.claude/agents/archon-<phase>.md` with per-phase `model:` frontmatter from `cfg.Models.Phases`.
2. Rewrite the shared "Phase Models" template to point the leader at the named `archon-<phase>` subagents and frame it as a hard gate (stop calling it advisory).
3. Fix opencode AGENTS.md wording (gate already exists in opencode.json).
- Fits as the claude equivalent of opencode-phase-subagents (Slice 2 of structured-model-resolution initiative).

## Completed Phases
- explore — completed 2026-06-27T06:00:00Z — recommendation: Approach C (per-phase `.claude/agents/archon-*.md` writer + Phase Models template rewrite). Diagnosis verified against real code.
- propose — completed 2026-06-27T06:20:00Z — `proposal.md` written. New capability `claude-phase-subagents`. ~250-320 lines, single PR under 400-line budget. 3 assumptions CONFIRMED by user (agent body follows sdd-<phase> skill; doc fix both harnesses no opencode behavior change; positioned as claude-parity of Writers — note opencode S2 already shipped in PR #48).
- spec — completed 2026-06-27T06:40:00Z — `specs/claude-phase-subagents/spec.md` + `.feature` written (7 requirements, 11 scenarios), mirroring `opencode-phase-writers`.
- design — completed 2026-06-27T07:00:00Z — `design.md` written. 7 decisions. New: `claude_mode.go` (`writeClaudeAgents`/`WriteClaudeAgents`, one .md per phase, atomic per-file, []string paths for rollback), `claude_mode_test.go`. Modify: `templates.go` (split trailer → phaseModelsClaude/phaseModelsOpencode + delegation rule rewrite), `init.go` (claude gate + rollback), `tui/model.go` (claude branch), `templates_test.go`. Effort/variant intentionally dropped from claude frontmatter (no platform field).

- tasks — completed 2026-06-27T07:30:00Z — `tasks.md` written. 17 tasks / 5 phases (Foundation, Core Writer, Wiring, Template Split, Tests). Forecast: ~270-340 lines, Low risk, single PR to master, no chaining, no pre-apply decision needed.
- judge — completed 2026-06-27T08:50:00Z — JUDGMENT: APPROVED. Dual blind review, both judges APPROVED, 0 confirmed issues. Mutation + Playwright gates skipped (disabled). Applied 2 test-only hardening suggestions from the judges; suite green. Report: `judge-report.md`.
- verify — completed 2026-06-27T08:25:00Z — PASS WITH WARNINGS. 17/17 tasks, 13/13 scenarios compliant, go build/vet/full suite green (11 pkgs). One WARNING (now CLOSED): scenario "Undo removes the generated agent files" was covered transitively; added end-to-end test `TestRun_ClaudeUndoRemovesGeneratedAgentFiles` (Run(claude)→Cleanup→stat) 2026-06-27T08:35:00Z — passes, suite green. Report: `verify-report.md`.
- apply — completed 2026-06-27T08:05:00Z — all 17 tasks done. New: `internal/initcmd/claude_mode.go` (+`claude_mode_test.go`). Modified: `init.go`, `tui/model.go`, `templates.go` (split into orchestratorRulesClaude/Opencode + orchestratorTrailerHead + phaseModelsClaude/Opencode + orchestratorStateManagement), `templates_test.go`. Independent verify: go build ✅, go vet ✅, go test ./internal/initcmd/... ./internal/tui/... ✅ green. ~159 prod lines + 538 test lines.

## Scope expansion (user-approved 2026-06-27, during design review gate)
The agent-definition model gate only fires if the leader delegates to the NAMED `archon-<phase>` subagent. Current Rule 2 ("delegate to sdd-* sub-agent") names the skill, not the model-bound agent. Scope now also rewrites the delegation rule in BOTH CLAUDE.md and AGENTS.md to name `archon-<phase>` as the per-phase target (claude: no per-call model param). Spec updated: now 8 requirements / 13 scenarios (added "Orchestrator delegates phases by named subagent"). proposal/design updated accordingly.

## Key Artifacts
- `openspec/changes/phase-model-hard-gate/state.yaml`
- `openspec/changes/phase-model-hard-gate/exploration.md` (pending)

## Open Questions
- Whether claude agent files should ALSO pass the per-invocation `model` param (belt-and-suspenders) or rely solely on frontmatter.
- Whether to mention `CLAUDE_CODE_SUBAGENT_MODEL` / `availableModels` caveats in generated docs.

## Next Recommended Step
Delegate exploration to sdd-explore sub-agent, then run Human Review Gate on exploration.md.
