# Proposal: Impeccable Frontend-Integration (design-language gate)

Builds on [exploration](exploration.md). Implements [[impeccable-gate]].

## Intent

Archon can gate web E2E behavior (Playwright) but has no design-language quality gate.
Target frontends ship inconsistent visuals/UX with no automated check. Integrate
[Impeccable](https://github.com/pbakaus/impeccable) as an opt-in, config-gated external
npm tool (`npx impeccable`) that runs a detection gate after verify/judge — mirroring the
Playwright precedent so reviewers and users get a familiar shape.

## Scope

### In Scope
- `Impeccable` config struct + `Config.Impeccable` field + `Clone()` copy (fields:
  `Enabled`, `AutoInstall`, `ProductPath`, `DesignPath`, `Severity`).
- `--impeccable` init flag, CLI get/set keys, TUI tab, `archon status` block.
- Thin `skills/impeccable/SKILL.md` orchestrating WHEN/HOW to call Impeccable (no detector
  reimplementation).
- Phase hooks: design (reference PRODUCT.md/DESIGN.md), apply (run passes), verify
  (advisory presence check), judge (detection gate), tasks/explore/spec selection.
- Preflight group F + orchestrator rule; repo `CLAUDE.md` mirror; README note.

### Out of Scope
- Vendoring Impeccable into the Go binary; reimplementing its 58-rule detector or LLM critique.
- A shared generic "web quality gates" abstraction (premature; keeps one-struct-per-gate).
- Auto-installing Impeccable by default (see decision 4).

## Capabilities

> Contract with sdd-spec.

### New Capabilities
- `impeccable-gate`: opt-in design-language integration — config surface, `--impeccable`
  flag/TUI/status, thin orchestration skill, and the post-verify/judge detection gate
  (blocking vs advisory semantics, Node-missing → blocked).

### Modified Capabilities
- None. Playwright/judge/phase behavior is unchanged when `impeccable.enabled: false`
  (zero-value default). Hooks are additive and flag-gated.

## Approach

Exploration approach **1** (full Playwright-mirror + thin skill). Replicate every mapped
Playwright touchpoint mechanically; author the two net-new surfaces (design-phase hook,
judge detection gate) fresh. Behavior is inert until the flag is on.

### Resolved Design Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | Deterministic 58-rule detector **BLOCKS** judge; LLM critique **ADVISORY** (report-only). `Severity` config knob (`block-deterministic` default) allows override. | Non-deterministic LLM checks would make the gate flaky; deterministic rules are safe to block. |
| 2 | Node/npx missing → judge returns **`blocked`** with actionable message, never silent pass. | Mirrors harness-judge "Playwright not installed → blocked" edge case. |
| 3 | `impeccable init` PRODUCT.md/DESIGN.md stay at **target-project root** (Impeccable default). Design phase **references** them; never overwrites SDD `design.md`. | Avoids name clash with `openspec/changes/.../design.md`; keeps ownership clear. |
| 4 | `impeccable.auto_install` **defaults OFF** — assume installed; when missing, **instruct** rather than silent `npx install`. | Matches "opt-in, no surprise side effects" ethos; install mutates the target repo. |
| 5 | Design-phase hook **included** but flagged highest-uncertainty — user may defer to a later slice. | Net-new surface with no Playwright precedent; lightweight reference keeps risk low. |

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/config/config.go` + `config_test.go` | Modified | `Impeccable` struct, field, `Clone()`, CloneRoundtrip fixture |
| `cmd/archon/config.go` | Modified | get/set keys + both key-list error strings |
| `cmd/archon/main.go`, `internal/initcmd/init.go` | Modified | `--impeccable` flag threaded through `buildConfig` |
| `internal/tui/impeccable_tab.go` (new) + `model.go` + `model_test.go` | New/Modified | tab cloned from playwright_tab; enum/label/routing/reload/save wiring |
| `internal/status/display.go` | Modified | Impeccable status block |
| `skills/impeccable/SKILL.md` (new) | New | thin orchestration skill; auto-embedded, bumps skill_count 24→25 |
| `skills/sdd-{design,apply,verify,tasks,spec,explore}/SKILL.md` | Modified | flag-gated phase hooks |
| `skills/harness-judge/SKILL.md` | Modified | Step 3c detection gate + result column + edge cases + output contract |
| `internal/initcmd/templates.go` + `templates_test.go` | Modified | preflight group F + rule; keep Claude/Opencode consts in sync |
| `CLAUDE.md`, `README.md`, `AGENTS.md` | Modified | mirror template (group F, rule); doc `--impeccable` |

## PR Slicing Plan

Session strategy **ask-always**, budget **400 lines**. Chained PRs, each < 400:

| PR | Scope | Est. lines |
|----|-------|-----------|
| **PR1 — Go wiring** | config struct/Clone/fixture, CLI get/set, `--impeccable` flag, `impeccable_tab.go` + model wiring + tests, status block. No behavior change. | ~330 |
| **PR2 — Orchestration** | `skills/impeccable/SKILL.md`, design/apply/verify/tasks/spec/explore hooks, templates group F + rule, CLAUDE.md mirror, README. | ~360 |
| **PR3 (conditional)** | Split judge detection gate + template group F out of PR2 if PR2 exceeds 400. | ~150 |

PR2 prose risk is real; expect to trigger PR3. Confirm the split at the ask-always gate.

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Flaky judge gate from LLM critique | Med | Decision 1: only deterministic rules block; `Severity` knob |
| Node/npx absent in target repo | High | Decision 2: `blocked` with actionable message |
| PRODUCT.md/DESIGN.md clash with SDD design.md | Med | Decision 3: target root, reference not overwrite |
| templates.go / repo CLAUDE.md drift | Med | Update both; templates_test.go catches template side |
| TUI tab index shift breaks tests | Med | Move `tabs` slice + index tests in lockstep with enum |
| Design-phase hook (no precedent) over-scoped | Med | Decision 5: lightweight, deferrable slice |
| PR2 exceeds 400-line budget | High | Pre-planned PR3 split; confirm at ask-always gate |

## Rollback Plan

Per-PR revert. Because behavior is fully gated on `impeccable.enabled` (zero-value
`false`), reverting any single PR — or leaving the flag off — restores prior behavior with
no migration. The judge/Playwright paths are untouched when disabled.

## Dependencies

- Node.js + `npx` in target projects (runtime, opt-in only; not a build dependency of Archon).
- Impeccable npm package (external; invoked, not vendored).

## Success Criteria

- [ ] `go test ./...` passes with the new `Impeccable` config, tab, and CloneRoundtrip fixture.
- [ ] `archon init --impeccable` sets `impeccable.enabled: true`; `archon status` shows the block; TUI tab toggles it.
- [ ] With flag off, judge and all phases behave identically to today (no new gate output).
- [ ] With flag on + Node present: deterministic detector failures block judge; LLM critique reports advisory.
- [ ] With flag on + Node absent: judge gate returns `blocked` with an actionable install/disable message.
- [ ] Each PR lands under the 400-line review budget.
