# Design: Proportional Phase Gating

<!-- [[harness-fast-track]] · [[harness-workflow]] · [[harness-judge]] ·
     [proposal](proposal.md) · [exploration](exploration.md) -->

## Technical Approach

Two additive seams on the existing track-as-data machinery
(`internal/workflow/workflow.go`): (1) a `fast` track value copying
`BugfixTrack`'s shape (workflow.go:30); (2) a new `Size` field on `State`
(workflow.go:58-64) persisting `sdd-explore`'s size classification so a
later, possibly-fresh `harness-judge` session can compute the risk tier. No
new Go config schema — `config.Judge` (config.go:23-25) stays `Enabled
bool`; tiering is Step 0 prose reading persisted state.

## Architecture Decisions

### Decision: Persist size on `state.yaml` via a new `Size` field + `--size` flag

**Choice**: Add `Size string \`yaml:"size,omitempty"\`` to `State` (workflow.go:58-64),
settable only via a new `--size` flag on `archon workflow complete`
(main.go:787-826) — the natural site, since size is a result `sdd-explore`
reports at explore's completion, mirroring how `Track` is set at bootstrap
via `advance --track`. Readable via `archon workflow status --json`
(main.go:663-676, add a `"size"` key) by the preflight suggestion step and
`harness-judge` Step 0.

**Alternatives considered**: (a) a separate file (`size.yaml` or
`exploration.md` frontmatter); (b) live orchestrator context only, never
persisted.

**Rationale**: (a) duplicates `state.yaml`'s tested atomic-write/`PreToolUse`
contract (`internal/initcmd/claude_hooks.go`) for one scalar; `exploration.md`
is free-form prose, not reliably parseable. (b) repeats a bug already fixed
once: `RetryCount`'s doc comment (workflow.go:54-57) exists "so it survives a
context compaction or a closed-and-reopened session instead of living only
in the orchestrating model's own count-as-you-go reasoning" — `harness-judge`
Step 0 must work in a fresh session with no memory of explore, so size needs
that same durable-state treatment, not in-context-only state.

### Decision: `fast` gets its own named slice, not a shared reference to `BugfixTrack`

**Choice**: `var FastTrack = []string{"explore","spec","apply","verify","archive"}`
(same phase names as `BugfixTrack`, workflow.go:30), own `case "fast":` in
`trackOrder()` (workflow.go:36-45).

**Alternatives considered**: returning `BugfixTrack` directly for `"fast"` too.

**Rationale**: phase *names* coincide today, but `fast`'s `spec` phase
produces a full Gherkin spec while `bugfix`'s produces the scoped 3-section
format (`sdd-spec` branches on `track`, not slice identity) — sharing the
slice would make a future divergence a surprising shared-reference edit.

### Decision: No `config.Judge` schema change

**Choice**: `judge.enabled` stays `Enabled bool` only; the computed tier
lives entirely in `harness-judge/SKILL.md` Step 0 prose, reading
`track`+`size` from `archon workflow status --json`.

**Rationale**: the spec requires a *computed* tier, not a *configurable*
toggle — no requirement asks for an opt-out of tiering, so a new Go field
would be scope the spec doesn't call for.

## Data Flow

    sdd-explore (Step 3c, classifies size)
              │
              ▼
    archon workflow complete --size {s}  →  state.yaml: size: {s}
              │
              ├──→ preflight Deviation rule (bullet d): reads status --json
              │     .size → suggests track:fast (small only, one question)
              │
              └──→ harness-judge Step 0: reads status --json .track + .size
                    → computes risk tier

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/workflow/workflow.go` | Modify | `FastTrack` + `"fast"` case in `trackOrder()` (error message lists `fast`); `Size` field on `State` |
| `internal/workflow/workflow_test.go` | Modify | Fast-track transition tests, `Size` roundtrip test |
| `cmd/archon/main.go` | Modify | `--size` flag on `newWorkflowCompleteCmd`; `size` key in `status --json` |
| `cmd/archon/main_test.go` | Modify | CLI tests for `--size` |
| `skills/sdd-explore/SKILL.md` | Modify | New Step 3c (size classification), mirrors Step 3b (lines 79-99) |
| `skills/harness-judge/SKILL.md` | Modify | Step 0 reads computed tier before `judge.enabled` check |
| `skills/harness-workflow/SKILL.md` | Modify | Phase Sequence (17-24) + Step 1 (99-101) gain `fast` |
| `internal/initcmd/templates.go` | Modify | Deviation rule (57-61) bullet (d): size-small suggests `fast`; Rule 6 (155/170): judge now tier-computed |

## Interfaces / Contracts

```go
var FastTrack = []string{"explore", "spec", "apply", "verify", "archive"}

type State struct {
    Phase      string         `yaml:"phase"`
    Status     string         `yaml:"status"`
    Track      string         `yaml:"track,omitempty"`
    Size       string         `yaml:"size,omitempty"` // small|substantial|unknown
    RetryCount int            `yaml:"retry_count,omitempty"`
    History    []HistoryEntry `yaml:"history,omitempty"`
}
```

`archon workflow complete --change {name} --size {small|substantial|unknown}`;
invalid value → non-zero exit, state.yaml untouched. `status --json` gains
`"size"` alongside existing keys.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `FastTrack` + `trackOrder("fast")` | `TestCheckTransition_FastTrackHappyPath`, `_PhaseNotInFastTrack`, `_BootstrapFastTrack` — mirror existing `*BugfixTrack*` tests |
| Unit | `State.Size` roundtrip | `TestState_SizeReadWriteRoundtrip` — mirror `TestState_ReadWriteRoundtrip` |
| CLI | `--size` flag + `status --json` | `TestWorkflowComplete_SizeFlag`, `_RejectsInvalidSize`, `TestWorkflowStatus_JSONIncludesSize` — mirror `TestWorkflowAdvance_Bootstraps*` (`git show b3bb6d1`) |
| Prose | Sizing judgment, judge tiering, preflight wording | No E2E harness for markdown-skill behavior — enforced by LLM-followed prose + Go-level `Size`/`Track` fields, as `bugfix` is today (prose for *when*, Go for *legal transition*) |

## Migration / Rollout

Additive only. Absent `size` on existing `state.yaml` files reads as
`unknown`, same back-compat posture as absent `track` → `full`. No migration
required.

## Open Questions

- [ ] Confirm whether `skills/sdd-spec/SKILL.md` already branches on `track`
  generically enough to emit a full Gherkin spec for `fast` (not the
  `bugfix` 3-section format) with no edit, or needs an explicit `fast`
  branch beside its existing `bugfix` one.
