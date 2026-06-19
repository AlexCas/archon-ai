# Verify + Judge Report — opencode-phase-writers (Slice 2)

- Branch: `feat/structured-models-writers`
- State: uncommitted working tree
- Reviewer mode: independent verify + adversarial judge (no code modified)
- Date: 2026-06-19

## VERDICT: SHIP

All gates pass, every spec scenario is satisfied and test-covered, and the adversarial
sweep (determinism, no-op correctness, FullID reuse, preservation, signature ripple,
scope, Go correctness) found no defects. Two LOW notes below, neither blocking.

---

## PART 1 — VERIFY (gates)

| Gate | Command | Result |
|------|---------|--------|
| Build | `go build ./...` | exit 0, no output |
| Vet | `go vet ./...` | exit 0, no output |
| Format | `gofmt -l` on the 5 changed files | exit 0, no files listed (clean) |
| Tests | `go clean -testcache && go test ./... -count=1` | exit 0, all 11 packages `ok` |

Test package summary: `cmd/archon`, `internal/agent`, `internal/config`,
`internal/initcmd`, `internal/models`, `internal/opencode`, `internal/scaffold`,
`internal/status`, `internal/tui`, `internal/version`, `skills` — all `ok`.

### Traceability table (scenario → code → test)

| Spec scenario | Code | Test | Status |
|---|---|---|---|
| Init writes a subagent per resolvable phase (no archon-judge) | `opencode_mode.go:91-99` loop over `ResolvePhaseModels`; judge absent because `PhaseOrder` excludes it (`config/model.go:167`) | `TestMergeOpencodeAgent_WritesSubagentPerResolvablePhase` | COVERED |
| A phase with no resolvable model is omitted (leader set, default+phases empty) | `ResolvePhaseModels` `continue` on empty (`config/model.go:257`); loop writes nothing | Covered by `_NothingConfiguredWritesNothing` (no default → no subagents) and `_PreservesExisting`/`_PhaseFallsBackToDefault` (default present → subagents). Leader-set-only-no-default exact pair not isolated — see LOW-1 | COVERED (indirect) |
| Subagent model == resolved FullID (`opencode/deepseek-v4-pro`) | `Model: pm.Model` (`opencode_mode.go:95`), `pm.Model == ref.FullID()` (`config/model.go:260`) | `TestMergeOpencodeAgent_PhaseModelMatchesResolvedFullID` | COVERED |
| Phase falls back to default model | `ResolvePhaseModels` `ref = mc.Default` (`config/model.go:254-255`) | `TestMergeOpencodeAgent_PhaseFallsBackToDefault` | COVERED |
| Fixed subagent shape (mode/hidden/description/prompt + order) | `archonPhaseAgent` struct + literal (`opencode_mode.go:32-38, 92-98`) | `TestMergeOpencodeAgent_SubagentFixedFields` + empirical JSON inspection | COVERED |
| Nothing configured writes nothing | `if leaderFull == "" && len(phases) == 0 { return "", nil }` (`opencode_mode.go:64-66`) | `TestMergeOpencodeAgent_NothingConfiguredWritesNothing` | COVERED |
| Phases set + empty leader → subagents, no leader | leader gated on `leaderFull != ""` (`opencode_mode.go:83`) | `TestMergeOpencodeAgent_PhasesSetEmptyLeaderWritesSubagentsNoLeader` (+ `_WritesSubagentPerResolvablePhase` asserts archon-leader absent) | COVERED |
| Re-run byte-identical + preserve user agents/top-level keys | sorted map keys via `MarshalIndent`, fixed struct order, trailing newline, atomic temp+rename (`opencode_mode.go:100-116`); additive merge preserves `doc` (`:79-100`) | `TestMergeOpencodeAgent_Idempotent` (now covers leader+all subagents) + `TestMergeOpencodeAgent_PreservesExisting` (user `agent.build`, `$schema`, `theme` survive; all subagents added) | COVERED |
| Leader shape/behavior unchanged | `archonLeaderAgent` struct + literal unchanged (`opencode_mode.go:18-23, 84-89`) | `TestMergeOpencodeAgent_CreatesAgent` | COVERED |
| TUI save == init merge (byte-identical, shared writer) | `tui/model.go:334` passes `cfg.Models` to same `MergeOpencodeAgent` | `TestSaveConfig_OpencodeLeaderMatchesInitMerge` (leader-only config) | COVERED (subagent parity structural — see LOW-2) |

No scenario gaps.

---

## PART 2 — JUDGE (adversarial findings)

Empirically inspected real JSON output via throwaway tests (removed after).

### Determinism — PASS
- Struct field order (`mode, hidden, model, description, prompt`) matches emitted JSON order.
- `Hidden bool` has no `omitempty` → serializes as `"hidden": true` (confirmed in output).
- `MarshalIndent` sorts the `agent` map: `archon-apply, archon-archive, archon-design,
  archon-explore, archon-leader, archon-propose, archon-spec, archon-tasks, archon-verify`
  plus user agents — stable across runs.
- Trailing newline appended (`opencode_mode.go:106`); atomic temp+rename (`:108-114`).
- Re-run with BOTH leader and subagents present is byte-identical (`_Idempotent` asserts
  `bytes.Equal`; passes). No non-determinism from Go map iteration because output goes
  through `MarshalIndent` key-sorting, not raw iteration.

### No-op correctness — PASS
- leader empty + phases present → subagents written, NO `archon-leader` key (the leader
  block is gated; it is not an empty-model leader). Confirmed.
- leader present + phases empty → only `archon-leader` written (empirically confirmed:
  leader-only output contained just `archon-leader`).
- both empty → early `return "", nil` BEFORE any read/write. Verified a pre-existing
  `opencode.json` is left byte-for-byte untouched (not truncated, not rewritten); a
  non-existent file is not created.

### ResolvePhaseModels reuse — PASS
- `pm.Model` is already `ref.FullID()` (`config/model.go:260`); writer assigns it verbatim
  (`Model: pm.Model`) — no double-join, no bare-alias leakage. `archon-spec` showed
  `opencode/deepseek-v4-pro` (provider-qualified) and fallback phases showed
  `anthropic/claude-sonnet-4-6`.
- `judge` never appears: `PhaseOrder` excludes it; test asserts `archon-judge` absent.
- A phase resolving only via Default IS written (`_PhaseFallsBackToDefault`; also every
  non-spec phase in the inspection used the default).

### Preservation — PASS
- Pre-existing user `agent.build` and top-level `$schema`/`theme` survive
  (`_PreservesExisting`). Only `archon-*` keys are set.
- An existing `archon-<phase>` from a prior run is overwritten via map key assignment, not
  duplicated (map keys are unique; `_Idempotent` byte-equality confirms no growth).

### Signature ripple — PASS
- All callers updated: `init.go:102`, `tui/model.go:334`, the 6 in-package test call sites,
  and `model_test.go:702` reference merge. `grep` for `mergeOpencodeAgent|MergeOpencodeAgent`
  shows no remaining string-argument callers.
- `nil`/zero `ModelConfig` handled: `config.ModelConfig{}` (nil `Phases` map) → `FullID()`
  on a zero `ModelRef` returns `""`, `ResolvePhaseModels` ranges a nil map safely (zero
  iterations) → clean no-op. `_NothingConfiguredWritesNothing` exercises this.

### Scope — PASS
- Diff touches only the 5 in-scope files (plus unrelated `CLAUDE.md`, not part of this slice).
- `internal/config/model.go` (ResolvePhaseModels, ModelRef, PhaseOrder), templates,
  scaffold, leader struct — all UNCHANGED. No variants/effort handling, no TUI picker.

### Go correctness — PASS
- Map mutation: `agents` reused when `doc["agent"]` is a map, else fresh map; reassigned to
  `doc["agent"]` (`:100`) — safe.
- Type assertions: `doc["agent"].(map[string]any)` uses comma-ok; on failure builds a fresh
  map rather than panicking. (Edge: if a user set `agent` to a non-object scalar, the merge
  silently replaces it — pre-existing behavior, unchanged by this slice.)
- Error wrapping: all `fmt.Errorf(..., %w, err)` — consistent with package convention.
- `.tmp` cleanup on rename failure: NOT removed on `os.Rename` error — this is the
  PRE-EXISTING behavior (identical before/after this change); not a regression. Noted only.

---

## LOW notes (non-blocking)

- LOW-1 (test coverage, not a code defect): the "@edge A phase with no resolvable model is
  omitted" scenario as literally written (leader set, default empty, all phases empty →
  zero subagents, leader present) has no single dedicated test. It is covered indirectly:
  the omit-when-empty path is exercised by `_NothingConfiguredWritesNothing`, and the leader
  emission by `_CreatesAgent`. The exact combined pair is not isolated. Consider adding one
  small test for completeness.
- LOW-2 (test coverage): `TestSaveConfig_OpencodeLeaderMatchesInitMerge` exercises only a
  leader-only config, so TUI/init byte-parity for subagents is structurally guaranteed
  (same shared writer, same `cfg.Models`) but not explicitly asserted. A models config with
  Default/Phases in that parity test would close the gap.

Neither note affects shippability. The implementation is correct, deterministic, idempotent,
in-scope, and faithful to the design and spec.
