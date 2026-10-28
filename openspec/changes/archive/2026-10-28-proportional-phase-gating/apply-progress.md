# Apply Progress: proportional-phase-gating

**Mode**: Standard (strict TDD disabled — `openspec/config.yaml` → `rules.apply.tdd: false`)
**Batch**: 1 of 1 (single-PR change, all 23 tasks across Phases 1-5 in this batch)

## Completed Tasks

### Phase 1: Foundation
- [x] 1.1 `internal/workflow/workflow.go`: added `var FastTrack = []string{"explore","spec","apply","verify","archive"}` as its own slice (not a `BugfixTrack` alias)
- [x] 1.2 `internal/workflow/workflow.go`: added `case "fast": return FastTrack, nil` in `trackOrder()`; error message now lists `fast`
- [x] 1.3 `internal/workflow/workflow.go`: added `Size string \`yaml:"size,omitempty"\`` to `State`
- [x] 1.4 `skills/sdd-spec/SKILL.md` Step 1b: `track: fast` now explicitly routes to the full Gherkin spec flow (Steps 2-6) alongside `full`/absent

### Phase 2: Core Implementation
- [x] 2.1 `cmd/archon/main.go`: `--size` flag on `newWorkflowCompleteCmd`, validated against `small|substantial|unknown`; invalid value rejected before any state read/write
- [x] 2.2 `cmd/archon/main.go`: `--size` wired into `state.Size` before `state.Write` (only when explicitly passed; empty string leaves existing `Size` untouched)
- [x] 2.3 `cmd/archon/main.go`: added `"size"` key to `status --json` output
- [x] 2.4 `cmd/archon/main.go`: extended `--track` help on `newWorkflowAdvanceCmd` to list `fast`

### Phase 3: Integration / Wiring
- [x] 3.1 `skills/sdd-explore/SKILL.md`: added Step 3c (size classification `small`/`substantial`/`unknown`), mirroring Step 3b's structure
- [x] 3.2 `skills/harness-judge/SKILL.md` Step 0: split into Step 0a (compute risk tier from `track`+`size` via `archon workflow status --json`) and Step 0b (the `judge.enabled` flag check); `judge.enabled: false` still always wins and short-circuits regardless of the computed tier
- [x] 3.3 `skills/harness-workflow/SKILL.md` "Phase Sequence": added `fast` bullet (`explore → spec → apply → verify → archive`, no judge)
- [x] 3.4 `skills/harness-workflow/SKILL.md` Step 1: added `track: fast` to the resolved-as-is and supported-tracks lists; also updated the Hard Rules bullet naming bugfix-only exemption to include `fast`
- [x] 3.5 `internal/initcmd/templates.go` Deviation rule: added bullet (d) — `size: small` suggests `track: fast` via one scoped question, never auto-selected
- [x] 3.6 `internal/initcmd/templates.go` Rule 6 (both Rules blocks — Claude and opencode): reworded to describe the tier-computed judge gate instead of an unconditional pass

### Phase 4: Testing
- [x] 4.1 `internal/workflow/workflow_test.go`: `TestCheckTransition_FastTrackHappyPath`
- [x] 4.2 `internal/workflow/workflow_test.go`: `TestCheckTransition_PhaseNotInFastTrack`
- [x] 4.3 `internal/workflow/workflow_test.go`: `TestCheckTransition_BootstrapFastTrack`
- [x] 4.4 `internal/workflow/workflow_test.go`: `TestCheckTransition_FastTrackVerifyToArchive`
- [x] 4.5 `internal/workflow/workflow_test.go`: `TestState_SizeReadWriteRoundtrip`
- [x] 4.6 `cmd/archon/main_test.go`: `TestWorkflowComplete_SizeFlag` + `TestWorkflowComplete_RejectsInvalidSize`
- [x] 4.7 `cmd/archon/main_test.go`: `TestWorkflowStatus_JSONIncludesSize`
- [x] 4.8 `go build ./...`, `go vet ./...`, `go test ./...` — all pass (no output from build/vet; all test packages `ok`)

### Phase 5: Cleanup
- [x] 5.1 Re-read `harness-judge/SKILL.md` + `harness-workflow/SKILL.md`: confirmed `judge.enabled: false` always-wins override and the single-judge (no dual review) guarantee both read unchanged after the tiering edits

## Files Changed

| File | Action | What Was Done |
|------|--------|---------------|
| `internal/workflow/workflow.go` | Modified | Added `FastTrack`, `"fast"` case in `trackOrder()` + updated error message, `Size` field on `State` |
| `internal/workflow/workflow_test.go` | Modified | 5 new tests: fast-track happy path, phase-not-in-fast-track, bootstrap-fast-track, verify-to-archive, `Size` roundtrip |
| `cmd/archon/main.go` | Modified | `--size` flag + validation on `workflow complete`, `"size"` key in `status --json`, `--track` help text lists `fast` |
| `cmd/archon/main_test.go` | Modified | 3 new tests: `TestWorkflowComplete_SizeFlag`, `TestWorkflowComplete_RejectsInvalidSize`, `TestWorkflowStatus_JSONIncludesSize`; added `encoding/json` import |
| `skills/sdd-explore/SKILL.md` | Modified | New Step 3c (size classification), mirrors Step 3b |
| `skills/sdd-spec/SKILL.md` | Modified | Step 1b routes `track: fast` to the full spec flow explicitly |
| `skills/harness-judge/SKILL.md` | Modified | Step 0 split into 0a (computed tier) and 0b (enabled flag, still wins); Hard Rules and Activation Contract reworded to match |
| `skills/harness-workflow/SKILL.md` | Modified | Phase Sequence gains `fast` bullet; Step 1 track-resolution list gains `fast`; bugfix-only Hard Rule now names `fast` too |
| `internal/initcmd/templates.go` | Modified | Deviation rule gains bullet (d) for size-small → suggest fast; Rule 6 in both Rules blocks (Claude + opencode) reworded for the tier-computed gate |

## Deviations from Design

None — implementation matches design.md. One clarification worth flagging: design.md's task-list wording for 3.2 said "insert computed-tier logic... BEFORE the existing judge.enabled check" while also saying the `enabled: false` override "must still short-circuit everything else." I resolved this by restructuring Step 0 into two documented sub-steps — 0a (compute tier, placed textually/first in document order) then 0b (the `judge.enabled` check) — while keeping the actual decision semantics identical to before: `judge.enabled: false` still always wins regardless of what the tier says. This satisfies both the literal "before" instruction and the "always-wins override" requirement; the behavior matches `harness-judge/spec.md`'s delta exactly (verified against its scenarios).

## Issues Found

None.

## Remaining Tasks

None — all 23 tasks across Phases 1-5 are complete.

## Workload / PR Boundary

- Mode: single PR (per tasks.md's Review Workload Forecast: ~230-280 lines, Low risk, Chain strategy: size-exception, Decision needed before apply: No)
- Current work unit: Unit 1 — "Full change (Go + docs)" (the only work unit)
- Boundary: this apply batch is the entire change's implementation. Apply's scope ends here; the archive commit (spec merge, folder move, SESSION_STATUS.md move, `archon map`) is staged after judge passes and before the PR is opened — not part of this batch.
- Estimated review budget impact: within forecast; no scope growth beyond what design.md/tasks.md specified.

## Status

23/23 tasks complete. Ready for verify.
