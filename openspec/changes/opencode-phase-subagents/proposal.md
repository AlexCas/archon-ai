# Proposal: opencode per-phase subagents in opencode.json

## Intent

`archon init` (and the TUI save path) only write a single primary agent,
`archon-leader`, into `opencode.json`. The per-phase models configured in
`.archon/config.yaml` (`cfg.Models.Phases`) never reach opencode, so opencode
runs every SDD phase on the leader's model and the per-phase configuration is
inert. We need opencode to actually honor the configured per-phase models.

## Scope

### In Scope
- Emit one subagent per SDD phase, keyed `archon-<phase>`, for the 8 phases in
  `config.PhaseOrder` (explore, propose, spec, design, tasks, apply, verify,
  archive).
- Each subagent carries its resolved per-phase model written verbatim.
- Extend the existing writer `mergeOpencodeAgent` (single atomic write that owns
  the whole `agent` map); update its exported seam and all callers.

### Out of Scope
- Multi-profile / model-profile support.
- A standalone `sync` command.
- Any change to the `archon-leader` agent shape or behavior.
- Any change to `NormalizeModel` / `ResolvePhaseModels` (no new model-resolution
  package — verbatim only).
- The `judge` phase (intentionally excluded from `PhaseOrder`).

## Capabilities

### New Capabilities
- `opencode-phase-subagents`: writing per-SDD-phase subagents (each with its
  per-phase model) into `opencode.json` during init and TUI save.

### Modified Capabilities
- None. (No existing spec covers opencode agent writing; the leader writer is
  uncovered today.)

## Approach (confirmed decisions = requirements)

1. **Keys**: `archon-<phase>` for each phase in `config.PhaseOrder` (8 phases).
2. **Model value, verbatim** — no `NormalizeModel`. Per-phase resolution chain:
   `cfg.Models.Phases[phase]` → `cfg.Models.Default` → `cfg.Models.Leader`.
3. **Always present**: a phase with no own model STILL gets a subagent using the
   fallback model. All 8 phases are emitted whenever the merge runs.
4. **Shape**: fixed-order struct, `mode: "subagent"`, `hidden: true`, plus
   model/description/prompt — for deterministic, byte-identical output.
5. **Single writer**: extend `mergeOpencodeAgent`; its signature grows to accept
   the model config (leader + default + phases). No sibling writer.
6. **Whole-merge no-op** redefined: today no-ops when `leader == ""`. New no-op =
   `leader == "" AND default == "" AND all phases empty`.
7. **Determinism preserved**: `MarshalIndent` (sorted keys) + trailing newline +
   atomic temp-file rename; existing top-level keys and pre-existing user agents
   preserved; only `archon-*` entries set.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/initcmd/opencode_mode.go` | Modified | Extend `mergeOpencodeAgent` + exported `MergeOpencodeAgent`; add per-phase subagent struct; rework no-op condition |
| `internal/initcmd/init.go:101-109` | Modified | Pass full model config to the writer |
| `internal/tui/model.go:333-334` | Modified | Update TUI save call site for new signature |
| `internal/initcmd/opencode_mode_test.go` | Modified | Cover per-phase emission, fallbacks, idempotency, preservation |
| `internal/tui/model_test.go:702` | Modified | Update for new signature |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Verbatim model is not provider-qualified, opencode rejects it | Med | Mirrors the leader's existing verbatim contract; user-owned, unchanged behavior |
| Non-idempotent output breaks re-runs | Low | Reuse existing deterministic machinery; golden-file idempotency test |
| Signature change misses a caller | Low | Grep-verified caller set; tests compile-gate all call sites |

## Rollback Plan

No new rollback wiring. The written `opencode.json` path is already registered
in `init.go` (`rollback.CreatedPaths`); extending the same writer is covered.
Revert is a single-commit `git revert`.

## Dependencies

None new. Reuses `internal/config` (`ModelConfig`, `PhaseOrder`) and existing
init/TUI wiring.

## Size / PR Forecast

Single PR, well under the 400-line budget (D1). Touches one writer plus its two
call sites and two test files — estimated ~150-250 changed lines, mostly tests.
No chained PR needed (C1 budget respected).

## Success Criteria

- [ ] After `archon init` for an opencode project, `opencode.json` contains
      `archon-<phase>` subagents for all 8 `PhaseOrder` phases, each with
      `mode: "subagent"`, `hidden: true`, and the verbatim resolved model.
- [ ] A phase with no own model uses the fallback model (default → leader).
- [ ] Re-running init or TUI save yields byte-identical `opencode.json`.
- [ ] Pre-existing top-level keys and user agents are preserved.
- [ ] Whole merge is a no-op only when leader, default, and all phases are empty.
