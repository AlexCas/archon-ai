# Exploration — opencode-phase-subagents

**Change**: Write a per-SDD-phase subagent (each with its per-phase model) into the
project's `opencode.json`, so opencode actually uses the configured per-phase models.
Today master only writes a single primary leader agent (`archon-leader`).

**Branch**: `feat/opencode-phase-subagents` (base `master` @ 7360b4d)
**Date**: 2026-06-19

## Context: what master already provides (reuse, don't duplicate)

- `internal/initcmd/opencode_mode.go` — `mergeOpencodeAgent` / `MergeOpencodeAgent`:
  reads `opencode.json` into a `map[string]any`, additively merges only
  `agent.archon-leader` (fixed-order struct `archonLeaderAgent`: `mode`/`description`/
  `model`/`prompt`), marshals with `MarshalIndent` (sorted keys) + trailing `\n`,
  atomic write via `.tmp` + `os.Rename`. No-op when `leader == ""`. Returns the written
  path for rollback. Callers: `init.go:101-109` (gated agent==opencode) and TUI save
  `model.go:333-334`.
- `internal/config/model.go` — `ModelConfig{Default, Leader, Phases}`;
  `PhaseOrder` (ordered, authoritative: explore, propose, spec, design, tasks, apply,
  verify, archive — **8 phases, `judge` excluded**); `ValidPhases` (same 8);
  `NormalizeModel(s)→(id,ok)`; `ResolvePhaseModels(mc)→[]PhaseModel`.
- `internal/models/` — catalog **enumerator only** (lists available models, strips
  `provider/` prefix). NOT a value normalizer. No provider-qualifier helper exists.
- `internal/initcmd/init.go` — Apply wiring + rollback registration
  (`rollback.CreatedPaths` appended before `WriteManifest`).

## Corrections to initial assumptions

- `config.ModelForPhase` **does not exist**. Real surface = `ResolvePhaseModels` +
  `NormalizeModel`.
- The SDD phase list is **8 delegated phases** (`config.PhaseOrder`), not 9 — `judge`
  is intentionally excluded.

## CENTRAL DESIGN PROBLEM — model string format mismatch

- opencode's `agent.<name>.model` expects a **provider-qualified id**
  (e.g. `anthropic/claude-sonnet-4-20250514`). Evidence: test fixture
  `testLeaderModel = "anthropic/claude-sonnet-4-20250514"`.
- The **leader** sidesteps this: archon writes `cfg.Models.Leader` **verbatim** — no
  normalization. Whatever the user typed lands as-is. (`models_tab.go:254`: "Written
  verbatim into opencode.json.")
- `ResolvePhaseModels`/`NormalizeModel` emit **bare aliases** (`opus`, `gpt-4o`) — the
  WRONG shape for opencode's `model` field. They feed the AGENTS.md advisory block, not
  opencode.json.

→ Implication: per-phase subagents should write the model **verbatim** from
  `cfg.Models.Phases[phase]` (fallback `cfg.Models.Default`), mirroring the leader's
  verbatim contract. Do NOT run them through `NormalizeModel` (would corrupt them into
  aliases opencode can't use). This also satisfies "reuse, don't add parallel
  resolution."

## opencode subagent JSON shape (gentle-ai reference pattern, ../gentle-ai)

`sdd-overlay-multi.json` per-phase subagent: `mode: "subagent"`, `hidden: true`,
`description`, `prompt`, `tools: {read,write,edit,bash}`. Keyed `sdd-<phase>`.
gentle-ai also stamps `variant` (effort) alongside `model`. Archon's leader is leaner
(4 fields, no tools/hidden/variant). Reference only — archon has no ModelAssignment/
variant/tools machinery.

## Key decisions for propose/design

1. **Subagent naming**: reuse `config.PhaseOrder`; key scheme `archon-<phase>` (matches
   `archon-leader`) vs gentle-ai's `sdd-<phase>`.
2. **Model value**: write `cfg.Models.Phases[phase]` verbatim, fallback `cfg.Models.Default`.
   Decide omit-when-empty vs always-emit.
3. **Subagent entry shape**: `mode: "subagent"`, `model`, `description`, `prompt`, and
   optionally `hidden`. Keep a fixed-order struct for deterministic JSON.
4. **Extend vs sibling**: EXTEND `mergeOpencodeAgent` (it already owns the whole `agent`
   map + single atomic write). Signature grows to accept phases (or full `ModelConfig`).
   Touches callers `init.go:102`, `model.go:334`, exported seam + tests.
5. **Rollback**: no change — same `opencode.json` path already registered.
6. **Idempotency / no-op**: preserve byte-identical re-runs; skip empty phases.
7. **Tests**: extend `opencode_mode_test.go`; update `tui/model_test.go:702` and call
   sites for the signature change.

## Key file:line references

- `internal/initcmd/opencode_mode.go:12,16-21,27,41-84`
- `internal/initcmd/init.go:101-109,232-236`
- `internal/config/model.go:9-13,82-96,148,179`
- `internal/models/{detect.go,opencode.go:36,resolve.go:24}`
- `internal/tui/model.go:333-334`; `internal/tui/models_tab.go:254,302`
- `internal/initcmd/opencode_mode_test.go:14,46,125,177,202,218`
- Reference: `../gentle-ai/internal/assets/opencode/sdd-overlay-multi.json:43-162`;
  `../gentle-ai/internal/components/sdd/inject.go:2189-2245`
