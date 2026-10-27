# Tasks: Proportional Phase Gating

<!-- [[harness-fast-track]] · [[harness-workflow]] · [[harness-judge]] ·
     [proposal](proposal.md) · [design](design.md) -->

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~230-280 |
| 800-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
800-line budget risk: Low

Well under the 800-line budget (precedent `b3bb6d1`: 170 lines, narrower
scope). Single PR, no slicing.

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Full change (Go + docs) | PR 1 | Single PR; base per session preflight |

## Phase 1: Foundation

- [ ] 1.1 `internal/workflow/workflow.go`: add `var FastTrack = []string{"explore","spec","apply","verify","archive"}` (own slice, not a `BugfixTrack` reference)
- [ ] 1.2 `internal/workflow/workflow.go`: add `case "fast": return FastTrack, nil` in `trackOrder()`; error message → `"unknown track: %s (supported: full, bugfix, fast)"`
- [ ] 1.3 `internal/workflow/workflow.go`: add `Size string \`yaml:"size,omitempty"\`` to `State`
- [ ] 1.4 `skills/sdd-spec/SKILL.md` Step 1b: edit routing so `track: fast` explicitly routes to Steps 2-6 (full Gherkin spec) alongside `full`/absent — resolves the open question

## Phase 2: Core Implementation

- [ ] 2.1 `cmd/archon/main.go`: `--size` flag on `newWorkflowCompleteCmd`, validated against `small|substantial|unknown`; invalid value → non-zero exit, state.yaml untouched
- [ ] 2.2 `cmd/archon/main.go`: wire `--size` into `state.Size` before `state.Write`
- [ ] 2.3 `cmd/archon/main.go`: add `"size"` key to `status --json` in `newWorkflowStatusCmd`
- [ ] 2.4 `cmd/archon/main.go`: extend `--track` help on `newWorkflowAdvanceCmd` to list `fast`

## Phase 3: Integration / Wiring

- [ ] 3.1 `skills/sdd-explore/SKILL.md`: add Step 3c (size classification `small`/`substantial`/`unknown`), mirroring Step 3b
- [ ] 3.2 `skills/harness-judge/SKILL.md` Step 0: insert computed-tier logic (reads `track`+`size` via `status --json`) before the `judge.enabled` check; preserve the `enabled: false` always-wins override exactly
- [ ] 3.3 `skills/harness-workflow/SKILL.md` "Phase Sequence" (17-24): add `fast` bullet (`explore → spec → apply → verify → archive`, no judge)
- [ ] 3.4 `skills/harness-workflow/SKILL.md` Step 1 (99-101): add `track: fast` to the resolved-as-is and supported-tracks lists
- [ ] 3.5 `internal/initcmd/templates.go` Deviation rule (~57-61): add bullet (d) — `size: small` suggests `track: fast` via one scoped question
- [ ] 3.6 `internal/initcmd/templates.go` Rule 6 (lines 155, 170, both Rules blocks): reword "After verify, invoke harness-judge" to describe the tier-computed gate

## Phase 4: Testing

- [ ] 4.1 `internal/workflow/workflow_test.go`: `TestCheckTransition_FastTrackHappyPath` — proves "Fast advances explore to spec"
- [ ] 4.2 `internal/workflow/workflow_test.go`: `TestCheckTransition_PhaseNotInFastTrack` — proves "Propose is not reachable in fast track"
- [ ] 4.3 `internal/workflow/workflow_test.go`: `TestCheckTransition_BootstrapFastTrack` — proves "Explicit fast track is honored"
- [ ] 4.4 `internal/workflow/workflow_test.go`: `TestCheckTransition_FastTrackVerifyToArchive` — proves "Verify transitions straight to archive regardless of tier"
- [ ] 4.5 `internal/workflow/workflow_test.go`: `TestState_SizeReadWriteRoundtrip` — mirrors `TestState_ReadWriteRoundtrip`
- [ ] 4.6 `cmd/archon/main_test.go`: `TestWorkflowComplete_SizeFlag` + `_RejectsInvalidSize`
- [ ] 4.7 `cmd/archon/main_test.go`: `TestWorkflowStatus_JSONIncludesSize` — roundtrips `--size` through `status --json`
- [ ] 4.8 Run `go build ./...`, `go vet ./...`, `go test ./...`; confirm all pass

## Phase 5: Cleanup

- [ ] 5.1 Re-read `harness-judge/SKILL.md` + `harness-workflow/SKILL.md`; confirm `judge.enabled: false` override and single-judge guarantee read unchanged
