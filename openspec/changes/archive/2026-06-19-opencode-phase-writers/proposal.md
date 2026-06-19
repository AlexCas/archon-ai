# Proposal — opencode-phase-writers (Slice 2)

**Slice 2 of `structured-model-resolution`.** Foundation (Slice 1) merged @ 7985c67.

## Problem

opencode reads per-agent models from `opencode.json` (`agent.<name>.model`). Today archon
writes only the single primary `archon-leader` agent there, so opencode runs every SDD phase
on the leader's model — the per-phase models configured in `.archon/config.yaml` never reach
opencode. The Foundation gave archon a structured `provider/model` (`ModelRef.FullID()`); this
slice uses it to actually write per-phase subagents.

## Goal

`archon init` and the TUI save path write one `archon-<phase>` subagent into `opencode.json`
per resolvable SDD phase, each carrying its resolved per-phase model (FullID), alongside the
existing `archon-leader`, by extending the single writer `mergeOpencodeAgent`.

## Scope

### In
- Extend `mergeOpencodeAgent`/`MergeOpencodeAgent` to also write `archon-<phase>` subagents.
- Per-phase model resolution REUSES `config.ResolvePhaseModels` (chain `Phases[phase]`→`Default`,
  omit-when-empty) — opencode subagents AGREE with the AGENTS.md "Phase Models" advisory.
- Update the two call sites and the writer's tests.

### Out (later/other slices)
- Provider→model TUI picker (Slice 3); effort/variants + opencode `variant` field (Slice 4).
- Any change to templates (already emit FullID post-foundation), `ResolvePhaseModels` semantics,
  or the leader's shape.

## Approach (settled decisions)

1. Keys `archon-<phase>` for the phases returned by `ResolvePhaseModels` (subset of `PhaseOrder`,
   judge excluded). A phase with no resolvable model is NOT written (consistent with the advisory).
2. Per-phase `model` = the `PhaseModel.Model` FullID from `ResolvePhaseModels`.
3. Per-phase entry shape (fixed field order): `mode:"subagent"`, `hidden:true`, `model`,
   `description` ("Archon SDD <phase> phase"), `prompt` ("{file:./AGENTS.md}").
4. Signature: `mergeOpencodeAgent`/`MergeOpencodeAgent` take `config.ModelConfig` (resolve leader
   + phases inside). Leader still written only when `Leader.FullID() != ""` (unchanged shape).
5. No-op when there is nothing to write: leader empty AND no resolvable phases.
6. Deterministic/idempotent (sorted keys, fixed struct order, trailing newline, atomic rename),
   preserve existing top-level keys + user agents. Rollback unchanged (same single file).

## Affected files
- `internal/initcmd/opencode_mode.go` (+ `opencode_mode_test.go`)
- `internal/initcmd/init.go:102` (pass `cfg.Models`)
- `internal/tui/model.go:334` (pass `cfg.Models`) (+ `internal/tui/model_test.go:675` ref merge)

## New capability to spec
- `opencode-phase-writers` — writing per-phase `archon-<phase>` subagents into opencode.json
  using ResolvePhaseModels FullIDs.

## Size / PR forecast
~150-250 LOC (prod ~40-60 + tests ~100-150). Under D1 400 — single PR. Chains off the merged
Foundation; base master.

## Risks
- Signature change misses a caller → compile-gated by tests (grep-verified: 2 call sites + tests).
- Behavior drift vs advisory → mitigated by REUSING ResolvePhaseModels (same resolver).
- Idempotency regression → golden byte-identical re-run test.
