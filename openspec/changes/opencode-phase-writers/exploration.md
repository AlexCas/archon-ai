# Exploration — opencode-phase-writers (Slice 2)

**Change**: Extend the opencode writer to emit per-SDD-phase `archon-<phase>` subagents
(+ leader) into `opencode.json`, each carrying its resolved per-phase model as a
`provider/model` FullID. Slice 2 of `structured-model-resolution`; Foundation (Slice 1)
merged to master @ 7985c67.

**Branch**: `feat/structured-models-writers` (base master @ 7985c67)
**Date**: 2026-06-19

## Current state (post-foundation)

- `internal/initcmd/opencode_mode.go` — writer UNCHANGED by foundation. Signatures:
  - `:27` `func MergeOpencodeAgent(projectDir, leader string) (written string, err error)`
  - `:41` `func mergeOpencodeAgent(projectDir, leader string) (...)`
  - `archonLeaderAgent{Mode, Description, Model, Prompt}` (`:16-21`, field order = JSON order).
  - No-op when `leader == ""` (`:42-44`). Additive merge, preserve existing, MarshalIndent
    sorted + trailing newline, atomic temp+rename, returns written path for rollback.
- Call sites already pass FullID (foundation S1e):
  - `init.go:102` `mergeOpencodeAgent(opts.ProjectDir, cfg.Models.Leader.FullID())`
  - `tui/model.go:334` `initcmd.MergeOpencodeAgent(m.projectDir, cfg.Models.Leader.FullID())`
- `config.ModelRef`/`FullID()`/`ParseModelRef`/`ModelConfig{Default,Leader ModelRef; Phases map[string]ModelRef}`/`PhaseOrder` (8 phases) all present.
- Templates ALREADY emit FullID via `ResolvePhaseModels` (`templates.go:162-171`, fed by
  `init.go:94` + `tui/model.go:354`). **Slice 2 does NOT touch templates.**
- Rollback: single `opencode.json` path registered at `init.go:106-108`; adding more agents
  to the same file needs NO new wiring.

## KEY DECISION — per-phase resolution chain (changed vs pre-pivot plan)

The foundation shipped `ResolvePhaseModels(mc)` (`model.go:250-263`) with the chain
**`Phases[phase]` → `Default`** (NO Leader fallback) and it **OMITS** a phase when neither
yields a model. The pre-pivot `opencode-phase-subagents` decision was "chain
Phases→Default→Leader, emit all 8 phases always." These now conflict.

- **Option A (reuse `ResolvePhaseModels`)**: emit a subagent only for phases that resolve
  (Phases→Default). A phase with no model is skipped. → the opencode subagents AGREE with the
  AGENTS.md "Phase Models" advisory block (same resolver). Smallest change, consistent.
- **Option B (new `Phases→Default→Leader` helper, always 8)**: matches the old decision but the
  Leader-tail fallback does not exist today and would make subagents DISAGREE with the advisory
  block (advisory omits an unresolved phase; subagent would carry the leader's model).

Recommendation: **Option A** — consistency between the advisory and the actual opencode agents
is valuable; the Leader-as-phase-fallback edge (default empty, leader set) is dubious value since
the primary `archon-leader` agent already covers that case.

## Slice 2 edits (design will finalize)

1. Signature: `mergeOpencodeAgent`/`MergeOpencodeAgent` take `config.ModelConfig` (resolve leader
   + phases inside). Update call sites `init.go:102`, `tui/model.go:334` to pass `cfg.Models`.
2. New `archonPhaseAgent` struct: `Mode("subagent")`, `Hidden(true)`, `Model`, `Description`
   ("Archon SDD <phase> phase"), `Prompt` ("{file:./AGENTS.md}"). Fixed field order.
3. Resolution: Option A reuse `ResolvePhaseModels` (pending user confirm).
4. No-op rework: nothing-to-write = leader FullID empty AND no resolvable phases. (Decide whether
   empty leader + present phases still writes the file — likely yes.)
5. Writer loop over resolved phases setting `agents["archon-"+phase]`; leader still gated on
   non-empty. Keep additive/atomic semantics.
6. Tests: update 6 `mergeOpencodeAgent(dir, testLeaderModel)` calls + the byte-identical TUI-vs-init
   reference test (`model_test.go:675`) to the new signature; add `phaseAgentFrom` helper; add
   subagent-shape + per-phase-model assertions; rework no-op test; extend preserves/idempotent.

## Size forecast
Prod ~30-50 LOC (struct + loop + no-op rework) + 2 one-line call sites; tests ~80-150 LOC.
Net ~150-250 LOC. Under D1 400 — single PR, no split.

## Reference (gentle-ai, pattern only)
`../gentle-ai/internal/assets/opencode/sdd-overlay-multi.json` per-phase: `mode:"subagent"`,
`hidden:true`, `description`, `prompt`, per-agent `model`. `injectModelAssignments`
(`inject.go:2189`) writes `model = assignment.FullID()` + `variant` (effort). archon Effort→variant
deferred to Slice 4.

## Key file:line
- `internal/initcmd/opencode_mode.go:12,16-21,27,41-84`
- `internal/initcmd/init.go:94,101-108`; `internal/tui/model.go:334,354`
- `internal/config/model.go:14-32,73-84,167,250-263`
- `internal/initcmd/templates.go:162-171,187`
- tests: `internal/initcmd/opencode_mode_test.go:14,17,31`; `internal/tui/model_test.go:675-713`
