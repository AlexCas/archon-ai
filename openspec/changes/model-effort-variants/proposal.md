# Proposal — model-effort-variants (Slice 4, option b)

**Optional final slice of structured-model-resolution.** Branch `feat/structured-models-effort`
(== master; Slices 1-3 merged).

## Problem
`config.ModelRef.Effort` exists but is never populated or written. opencode supports a per-agent
`variant` (effort/reasoning level), but archon never sets it, so users can't control reasoning
effort for models that support it.

## Goal
Let the user pick an effort level (default/low/medium/high) for reasoning-capable models in the TUI
picker, persist it on the `ModelRef`, and write it to opencode.json as `variant`. No opencode plugin,
no variants cache, no embedded-asset subsystem (option b — derive availability from the existing
`Model.Reasoning` flag).

## Scope
### In
- `PhaseModel` carries `Effort`; `ResolvePhaseModels` propagates `ref.Effort`.
- `opencode_mode.go`: `archonLeaderAgent` + `archonPhaseAgent` gain `Variant string json:"variant,omitempty"`;
  populated from the leader's / phase's resolved `Effort`.
- TUI picker: a new `effortSelect` sub-mode entered ONLY after picking a model whose
  `opencode.Model.Reasoning == true`; options `default`(→"")/`low`/`medium`/`high`; sets `row.ref.Effort`.
  Non-reasoning models skip the step (effort cleared).
- Tests across config / initcmd / tui.

### Out
- Per-model accurate variant lists (would need the opencode plugin + cache — option a, rejected).
- Free-form effort parsing (free-form path leaves Effort empty).
- Any change to the config YAML round-trip (MarshalYAML already handles `{provider,model,effort}`).

## Settled decisions
- Option (b): effort offered only for `Reasoning` models; fixed default/low/medium/high.
- `variant` written with `omitempty` — idempotency-safe because archon rewrites the whole `archon-*`
  entry each run (no stale variant survives). Confirmed by an idempotency test.
- Free-form entry does not set Effort.

## Affected files
- `internal/config/model.go` (+ model_test.go)
- `internal/initcmd/opencode_mode.go` (+ opencode_mode_test.go)
- `internal/tui/models_tab.go` (+ models_tab_test.go)

## New capability to spec
- `model-effort-variants` — effort selection for reasoning models + `variant` write.

## Size / PR forecast
~150-250 LOC (prod ~90-130 + tests ~80-130). Under D1 400. Single PR off master.

## Risks
- Idempotency with omitempty variant → covered by a re-run byte-identical test.
- Picker state-machine regression → effortSelect mirrors the existing sub-mode pattern; full test coverage.
- Back-compat: a ModelRef with Effort marshals as a mapping (already supported/tested) — no churn for
  effortless configs.
