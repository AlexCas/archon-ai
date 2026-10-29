## Verification Report

**Change**: proportional-phase-gating
**Version**: N/A (no version bump declared in proposal/design)
**Mode**: Standard (Strict TDD disabled — `openspec/config.yaml` → `rules.apply.tdd: false`, per `apply-progress.md`)

### Completeness

| Metric | Value |
|--------|-------|
| Tasks total | 23 |
| Tasks complete | 23 |
| Tasks incomplete | 0 |

All 23 checkboxes in [tasks.md](tasks.md) are `[x]`, matching `apply-progress.md`'s own
"23/23 tasks complete" status line. No cleanup or core tasks outstanding.

### Build & Tests Execution

**Build**: ✅ Passed
```text
$ go build ./...
(no output — clean build)
```

**Vet**: ✅ Passed
```text
$ go vet ./...
(no output — clean)
```

**Tests**: ✅ 15/15 packages passed, 0 failed, 0 skipped (full repo suite, not just the touched packages)
```text
$ go test -count=1 ./...
ok  	github.com/archon-ai/archon/cmd/archon
ok  	github.com/archon-ai/archon/internal/agent
ok  	github.com/archon-ai/archon/internal/config
ok  	github.com/archon-ai/archon/internal/initcmd
ok  	github.com/archon-ai/archon/internal/mapgen
ok  	github.com/archon-ai/archon/internal/models
ok  	github.com/archon-ai/archon/internal/opencode
ok  	github.com/archon-ai/archon/internal/route
ok  	github.com/archon-ai/archon/internal/scaffold
ok  	github.com/archon-ai/archon/internal/skillregistry
ok  	github.com/archon-ai/archon/internal/status
ok  	github.com/archon-ai/archon/internal/tui
ok  	github.com/archon-ai/archon/internal/version
ok  	github.com/archon-ai/archon/internal/workflow
ok  	github.com/archon-ai/archon/skills
```

**Named tests** (all 8 called out in the task brief, run individually with `-v`, confirmed passing, not just read):
```text
$ go test -count=1 -v ./internal/workflow/... -run 'TestCheckTransition_FastTrackHappyPath|TestCheckTransition_PhaseNotInFastTrack|TestCheckTransition_BootstrapFastTrack|TestCheckTransition_FastTrackVerifyToArchive|TestState_SizeReadWriteRoundtrip'
--- PASS: TestCheckTransition_PhaseNotInFastTrack (0.00s)
--- PASS: TestCheckTransition_FastTrackHappyPath (0.00s)
--- PASS: TestCheckTransition_FastTrackVerifyToArchive (0.00s)
--- PASS: TestCheckTransition_BootstrapFastTrack (0.00s)
--- PASS: TestState_SizeReadWriteRoundtrip (0.00s)
PASS

$ go test -count=1 -v ./cmd/archon/... -run 'TestWorkflowComplete_SizeFlag|TestWorkflowComplete_RejectsInvalidSize|TestWorkflowStatus_JSONIncludesSize'
--- PASS: TestWorkflowComplete_SizeFlag (0.00s)
--- PASS: TestWorkflowComplete_RejectsInvalidSize (0.00s)
--- PASS: TestWorkflowStatus_JSONIncludesSize (0.00s)
PASS
```

Read, not just trusted by name: each test body was inspected.
`TestCheckTransition_FastTrackHappyPath` walks `explore→spec→apply→verify→archive`
and asserts each hop is `Allowed`. `TestCheckTransition_PhaseNotInFastTrack` asserts
`propose` on a `fast`-track change is blocked with the exact reason string
`"propose is not part of the fast track"`. `TestCheckTransition_BootstrapFastTrack`
asserts a brand-new `fast` change may only start at `explore`.
`TestCheckTransition_FastTrackVerifyToArchive` asserts `verify→archive` is allowed on
`fast` with no judge hop. `TestState_SizeReadWriteRoundtrip` round-trips `Size: "small"`
through `Write`/`ReadState`. `TestWorkflowComplete_SizeFlag` bootstraps a change,
runs `workflow complete --size small`, and reads `state.yaml` back to confirm
`Size == "small"`. `TestWorkflowComplete_RejectsInvalidSize` passes `--size bogus`
and asserts `state.yaml` is byte-for-byte unchanged from before the call (not just
that the command errored). `TestWorkflowStatus_JSONIncludesSize` round-trips
`--size substantial` through `status --json` and asserts the `"size"` key is present
with that value.

**Coverage**: Not separately measured for this change; not available / not requested in design.md's Testing Strategy. ➖ Not available.

### Spec Compliance Matrix

#### `harness-fast-track.feature` (7 scenarios)

| Scenario | Test / Evidence | Result |
|----------|------------------|--------|
| Explicit fast track is honored | Go half (sequence resolves): `TestCheckTransition_BootstrapFastTrack` + `trackOrder("fast")`. Spec-format half (full Gherkin, not bugfix 3-section): `skills/sdd-spec/SKILL.md` Step 1b — `track: fast` routes to Steps 2-6 (source-inspected, verbatim in `git show 6dc3af3`) | ✅ COMPLIANT (test + inspection, split per design) |
| Fast advances explore to spec | `TestCheckTransition_FastTrackHappyPath` (first hop of the walked chain) | ✅ COMPLIANT |
| Propose is not reachable in fast track | `TestCheckTransition_PhaseNotInFastTrack` | ✅ COMPLIANT |
| Verify transitions straight to archive regardless of tier | `TestCheckTransition_FastTrackVerifyToArchive` | ✅ COMPLIANT |
| Small classification suggests fast via one scoped question | `skills/sdd-explore/SKILL.md` Step 3c + `internal/initcmd/templates.go` Deviation rule bullet (d) — prose only, no Go test possible | ⚠️ COMPLIANT (inspection) — see "Prose-only scenarios" below |
| Unknown classification defaults to full | `skills/sdd-explore/SKILL.md` Step 3c ("`unknown` → track defaults to `full`; `fast` is never suggested") + reinforced by the pre-existing Go default (`trackOrder("")`/absent → `full`, unchanged by this diff) | ⚠️ COMPLIANT (inspection) |
| Human Review Gate fires once after spec | CLAUDE.md Human Review Gate (unconditional, applies to any phase producing an editable artifact) + `harness-fast-track/spec.md`'s own requirement text — prose only | ⚠️ COMPLIANT (inspection) |

#### `harness-judge.feature` (5 scenarios)

| Scenario | Test / Evidence | Result |
|----------|------------------|--------|
| Explicit judge.enabled false always skips, regardless of tier | `skills/harness-judge/SKILL.md` Step 0b — `judge.enabled: false` branch explicitly short-circuits "not even when the Step 0a tier would have required judge" (unchanged wording from before tiering existed, confirmed by diff) | ⚠️ COMPLIANT (inspection) |
| Low-risk computed tier skips or lightens judge | Step 0b "Tier skip-eligible" branch | ⚠️ COMPLIANT (inspection) — **see Issue W1 below**: the "archive is reachable" half of this is not actually unconditional at the Go layer |
| Re-apply loop and retry cap are unchanged | Preserved verbatim (Steps 5/6); `internal/workflow.MaxJudgeRetries`/`ErrMaxRetriesExceeded` untouched by this diff | ⚠️ COMPLIANT (inspection, preserved not new) |
| Substantial full-track change always requires judge | Step 0a computation rule: `track: full` + `size` substantial/unknown → "requires judge" | ⚠️ COMPLIANT (inspection) |
| Small full-track change computes a skip-eligible tier | Step 0a computation rule: `track: full` + `size: small` → "skip-eligible" | ⚠️ COMPLIANT (inspection) |

#### `harness-workflow.feature` (6 scenarios)

| Scenario | Test / Evidence | Result |
|----------|------------------|--------|
| One judge gates archive when the tier requires review | `skills/harness-workflow/SKILL.md` Phase Sequence + `harness-judge` Step 0a "requires judge" path; unconditional-pass invariant was already Go-enforced (`config.PhaseOrder` places `judge` immediately before `archive`) | ✅ COMPLIANT (test [transition order] + inspection [tiering decision]) |
| No second judge is added to the SDD path | Unchanged single-judge delegation (`harness-judge` Hard Rules: "Do NOT run a dual or blind review") — preserved, not new | ⚠️ COMPLIANT (inspection, preserved not new) |
| Low-risk tier skips judge and archive proceeds from verify | `harness-judge` Step 0b skip branch | ⚠️ COMPLIANT (inspection) — **see Issue W1**: empirically, `archon workflow advance --to archive` from `verify/completed` on `track: full` is Go-blocked (`missing phases: judge`) unless the orchestrator still drives judge to `completed` first; the skip-path prose doesn't say so |
| Explicit judge.enabled false always skips, regardless of tier | Same as harness-judge.feature's identical scenario above | ⚠️ COMPLIANT (inspection) |
| Fast track is resolved alongside full and bugfix | `TestCheckTransition_BootstrapFastTrack` + `trackOrder("fast")` (Go) + `harness-workflow/SKILL.md` Phase Sequence/Step 1 listing `fast` (prose) | ✅ COMPLIANT (test + inspection) |
| Fast is never silently auto-selected | `skills/sdd-explore/SKILL.md` Step 3c + `templates.go` Deviation rule bullet (d): "MAY suggest... via one scoped question... MUST NOT silently auto-select" — prose only | ⚠️ COMPLIANT (inspection) |

**Compliance summary**: 18/18 scenarios compliant (6 by runtime test or test+inspection, 12 by source inspection of prose that has no possible Go-level test — see rationale below); 1 WARNING raised against an inspection-COMPLIANT scenario (archive-reachability mechanics, not a scenario-text violation).

### Prose-only scenarios — precedent-grounded treatment (not CRITICAL `UNTESTED`)

Several scenarios describe LLM-followed orchestration prose with no possible Go test:
the sizing judgment itself (`small`/`substantial`/`unknown`), the risk-tier computation
in `harness-judge` Step 0a, the "one scoped question, never silent" suggestion rule, the
Human Review Gate firing on the fast track, and the preserved re-apply loop/single-judge
invariants. `design.md`'s own Testing Strategy table flags this explicitly: *"No E2E
harness exists for markdown-skill-level behavior... enforced by LLM-followed prose +
Go-level Size/Track fields, as bugfix is today."*

Per the `sdd-verify` skill's Decision Gates table, a scenario with no covering test is
normally `CRITICAL UNTESTED`. I checked this project's actual precedent before applying
that rule here, rather than guessing:

1. **`openspec/specs/harness-bugfix-track/spec.md`** — the sibling track this change is
   explicitly modeled on — has scenarios in exactly this shape: *"Human Review Gate fires
   once after the scoped spec"*, *"judge.enabled does not add a judge to the bugfix
   track"*. These are equally prose-enforced orchestrator behavior with no dedicated Go
   test for the skill-level enforcement (only the Go-level phase-sequence mechanics are
   tested). The change that introduced `harness-bugfix-track`
   (`openspec/changes/archive/2026-09-16-streamline-harness/`) left no `verify-report.md`
   in its archived folder — its `state.yaml` stops at `design/completed`, suggesting it
   predates this project's current verify-report convention entirely, so it offers no
   directly comparable report, but it confirms the spec-authoring precedent of writing
   prose-only scenarios without a Go-test requirement.
2. **`openspec/changes/archive/2026-07-25-impeccable-frontend-integration/verify-report.md`**
   is the closest directly comparable case: a thin orchestration-prose feature (flag-gated
   skill hooks across `sdd-explore`/`sdd-spec`/`sdd-design`/`sdd-apply`/`sdd-verify`/
   `sdd-tasks`/`harness-judge`, explicitly "no Go-side detector," matching this change's own
   "no new Go config schema, tiering lives entirely in SKILL.md prose" design decision).
   Its verify-report maps each such scenario to a `COMPLIANT (inspection)` / `COVERED
   (prose)` status — e.g. *"Judge gate passes when no violations... COVERED (prose)"*,
   *"Skill delegates detection to npx... COMPLIANT (inspection)"* — and states the rationale
   verbatim: *"there is no Go test harness that executes `sdd-*` skill prose, so these
   scenarios are verified by source inspection of the skill text (the only possible
   evidence for prose hooks) rather than runtime tests."* Its final verdict was **PASS WITH
   WARNINGS**, with the warnings being unrelated issues (a stray file deletion) — not the
   prose-only scenarios themselves.

**My call**: I am treating this change's prose-only scenarios the same way —
`COMPLIANT (inspection)`, not `CRITICAL UNTESTED` — because (a) this is consistent,
repeated, documented project precedent for this exact style of skill-level behavior,
not a standard I am inventing; (b) each scenario has concrete, checkable inspection
evidence (specific SKILL.md lines, confirmed present in the actual file, not just
claimed in a commit message); and (c) applying `UNTESTED`/CRITICAL here would be a
*stricter* standard than this project has ever held this category of change to,
which the `sdd-verify` skill's own Decision Gates table does not require me to
invent — it says "no covering test" is CRITICAL, but inspection-backed prose that
literally cannot have a Go test is exactly the case this project's own prior verify
practice already carved out. A human reviewer who disagrees with this judgment call
should treat it as the single most consequential thing to double-check in this report.

### Correctness (Static Evidence)

| Requirement | Status | Notes |
|------------|--------|-------|
| Track field supports fast | ✅ Implemented | `trackOrder("fast")` returns `FastTrack`; error message lists `fast` |
| Fast phase sequence | ✅ Implemented | `FastTrack = []string{"explore","spec","apply","verify","archive"}`; confirmed via tests above |
| Judge absent in fast track | ✅ Implemented | No `judge` in `FastTrack`; `harness-judge` Step 0a explicitly states fast/bugfix never reach the gate |
| Suggestion-only entry via sizing classification | ✅ Implemented | `sdd-explore` Step 3c + `templates.go` bullet (d); no code path auto-writes `track: fast` |
| Human Review Gate preserved | ✅ Implemented (prose, unconditional CLAUDE.md gate applies regardless of track) | |
| Computed risk tier gates judge (full track only) | ✅ Implemented | Step 0a/0b split, confirmed non-contradictory (see Design Coherence) |
| Computed Per-Change Risk Tier | ✅ Implemented | Step 0a's 3-way rule matches spec exactly (`full`+substantial/unknown → requires; `full`+small → skip-eligible; non-full → N/A) |
| Single Judge in the Full-Track Terminal Ordering (MODIFIED) | ⚠️ Implemented with a mechanics gap | See Issue W1 |
| Fast Track Selection in Terminal Ordering (ADDED) | ✅ Implemented | |

### Coherence (Design)

| Decision | Followed? | Notes |
|----------|-----------|-------|
| `FastTrack` is its own named slice, not a `BugfixTrack` alias | ✅ Yes | `git show 6dc3af3`: `var FastTrack = []string{...}` is a distinct declaration with its own doc comment explaining why (sdd-spec branches on track, not slice identity) |
| `Size` field added to `State`, `yaml:"size,omitempty"` | ✅ Yes | `internal/workflow/workflow.go`; roundtrip test passes |
| No `config.Judge` schema change | ✅ Yes | `internal/config/config.go`'s `Judge` struct is still exactly `{Enabled bool}`; `git diff b3bb6d1 6dc3af3 -- internal/config/config.go` is empty — confirmed zero diff, not just "no judge-shaped diff" |
| `--size` flag validates before any write | ✅ Yes | `cmd/archon/main.go`: the `validWorkflowSizes` check runs and returns before `os.Getwd()`/state read/write; `TestWorkflowComplete_RejectsInvalidSize` asserts `state.yaml` is byte-identical to before the call |
| `sdd-spec` open question resolved (track:fast → full Gherkin spec) | ✅ Yes | Task 1.4 closed design.md's own Open Question; confirmed in `skills/sdd-spec/SKILL.md` Step 1b diff |
| `harness-judge` Step 0a/0b internal contradiction fix | ✅ Yes — re-verified in the current file, not just the commit message | See dedicated check below |

**Dedicated re-check: Step 0a/0b contradiction (the fix I was asked to confirm, not just trust).**
I re-read the current `skills/harness-judge/SKILL.md` directly (not the commit diff alone).
Step 0a says: *"`track` != `full` → tiering does not apply; judge is not run for this track
at all (governed by that track's own spec, not this gate)."* Step 0b's `judge.enabled: true`
branch now explicitly handles the only way this could still read as a contradiction — it adds:
*"the only track that reaches this skill at all — `bugfix` and `fast` have no judge phase in
their sequence, so the orchestrator never invokes `harness-judge` for them; if you somehow
got here on a non-full track, treat it as an orchestrator routing bug and report `blocked`
rather than guessing a tier."* This closes the gap: there is no remaining path where Step 0b
computes "requires judge" for a non-full track while Step 0a says tiering doesn't apply —
Step 0b now defers to Step 0a's framing explicitly instead of re-deciding independently.
**Confirmed fixed in the current file.**

### Issues Found

**CRITICAL**: None.

**WARNING**:
- **W1 — Skip-eligible-tier "archive reachable directly from verify" is not mechanically
  unconditional at the Go layer; the skip-path prose doesn't close the gap.** Empirically
  confirmed with the built binary: on `track: full`, with `state.yaml` at `phase: verify,
  status: completed`, running `archon workflow advance --to archive` is blocked —
  `"blocked: requested phase skips one or more mandatory phases" / "missing phases: judge"`
  — regardless of `size: small`/skip-eligible tier. `config.PhaseOrder` places `judge`
  immediately before `archive` and `internal/workflow.CheckTransition` was not (and per
  design.md, was never intended to be) modified to carve out an exception; tiering is pure
  orchestrator-prose. The only way to reach `archive` is still to drive `judge` through
  `in_progress` → `completed` first (confirmed working when I did this manually with the
  built binary). `harness-judge/SKILL.md` Step 0b's skip branches say "leave `state.yaml`
  for the orchestrator to advance to `archive`" without instructing that `archon workflow
  complete --change {name}` must still be called for the `judge` phase first — only Step 5
  ("On Pass") explicitly says to call it. **This exact gap is pre-existing, not a regression
  introduced by this change**: the quoted phrase is verbatim carried over, unchanged, from
  the `judge.enabled: false` branch that existed before this diff (confirmed via `git show
  6dc3af3 -- skills/harness-judge/SKILL.md` — that line is context, not a `+` addition), and
  the new skip-eligible-tier branch explicitly says it is "mirroring the format above." So
  this is an existing, inherited ambiguity that this change's new scenario text (*"archive is
  reachable directly from verify"*, harness-workflow and harness-fast-track specs) states
  more confidently than the actual mechanics support, without closing it. Recommend (not
  applying — out of scope for verify): add one sentence to both skip branches in Step 0b
  instructing `archon workflow complete --change {name}` for the `judge` phase (with the
  skip `Reason` as the completion note) before the orchestrator advances to `archive`.

  **Resolution (orchestrator, post-verify)**: independently reproduced the exact
  sequence with the built binary (`archon workflow advance --to judge` →
  `archon workflow complete` → `archon workflow advance --to archive`) and confirmed
  it succeeds cleanly — the Go layer was never actually blocking the skip path; it
  just requires the ordinary complete-then-advance step like every other phase,
  which the prose didn't spell out. Applied the recommended one-line clarification to
  both Step 0b skip branches in `skills/harness-judge/SKILL.md` (committed separately
  from this report). No Go-level change was needed or made.

**SUGGESTION**: None beyond W1's recommended one-line clarification, which is folded into W1 rather than listed separately.

### Verdict

**PASS WITH WARNINGS**

All 23 tasks complete; build, vet, and the full test suite (15 packages, including all
8 named tests) pass at runtime; every Gherkin scenario across all three `.feature` files
has either a passing covering test or documented, precedent-consistent inspection
evidence; all six design.md architecture decisions are followed, including the two I was
specifically asked to re-verify in the live file (no `config.Judge` schema change, Step
0a/0b contradiction fix). One WARNING (W1) is raised: a pre-existing (not newly
introduced) mechanical gap between the skip-path prose and the Go state machine's
`archive`-requires-`judge`-completed invariant, which the new tiering feature inherits
without closing and which this change's new scenario wording states more confidently
than current mechanics guarantee. Not CRITICAL — it does not block any task, build, or
test, and the underlying mechanism to satisfy it already exists (`workflow complete` on
`judge`); it is a documentation/clarity gap in orchestrator prose, worth a follow-up but
not a reason to fail this change's own deliverable.

---

## Re-Verification Report (2026-10-27) — post-judge-fix retry cycle

**Context**: the first `archon-judge` review (of commit `6dc3af3`'s implementation)
failed with 4 confirmed issues. Commit `11c19f5` fixed all 4. This section
independently re-verifies those fixes — it does not repeat the full original verify
pass above, which already passed and is unaffected by this commit's changes. Commit
`11c19f5` touched exactly: `cmd/archon/main.go`, `cmd/archon/main_test.go`,
`openspec/changes/proportional-phase-gating/{design.md,state.yaml}`,
`openspec/changes/proportional-phase-gating/specs/harness-fast-track/spec.md`,
`openspec/map.md`, `skills/harness-judge/SKILL.md`, `skills/harness-workflow/SKILL.md`,
and `skills/sdd-explore/SKILL.md` — nothing else in the original
compliance/correctness/coherence tables above is affected).

### Fix #1 — `track:fast` reachability (CRITICAL in original judge review)

**Code inspection** (`cmd/archon/main.go`, current `newWorkflowCompleteCmd`):
confirmed `--track` flag exists; validated against a `validWorkflowTracks` map of
exactly `full|bugfix|fast`, rejected with `"invalid --track: %s (supported: full,
bugfix, fast)"` before `state.yaml` is even read; once `state` is loaded, a second
guard rejects `trackFlag != "" && state.Phase != "explore"` with `"--track is only
accepted while completing the explore phase (current phase: %s)"` — both rejections
happen strictly before `state.Write()`. When valid, `state.Track = trackFlag` is
applied in the same call that also applies `--size` and `state.Complete()`.

**New tests run directly** (not trusted from the commit message):
```text
$ go test -count=1 -v ./cmd/archon/... -run 'TestWorkflowComplete_TrackSwitchAtExplore|TestWorkflowComplete_TrackRejectedAfterExplore|TestWorkflowComplete_RejectsInvalidTrack'
=== RUN   TestWorkflowComplete_TrackSwitchAtExplore
--- PASS: TestWorkflowComplete_TrackSwitchAtExplore (0.00s)
=== RUN   TestWorkflowComplete_TrackRejectedAfterExplore
--- PASS: TestWorkflowComplete_TrackRejectedAfterExplore (0.00s)
=== RUN   TestWorkflowComplete_RejectsInvalidTrack
--- PASS: TestWorkflowComplete_RejectsInvalidTrack (0.00s)
PASS
ok  	github.com/archon-ai/archon/cmd/archon	0.008s
```
All three pass. `TestWorkflowComplete_TrackSwitchAtExplore` additionally asserts the
switch takes effect for the *next* transition (`advance --to spec` succeeds, fast's
second phase, not `propose`). `TestWorkflowComplete_TrackRejectedAfterExplore` asserts
`state.yaml` is byte-equivalent (`Status`/`Track` unchanged) after a rejected call.

**Independent empirical check with the built binary** (`go install ./cmd/archon`,
exercised in a scratch temp dir, not the test suite):
```text
$ archon workflow advance --change my-change --to explore
my-change advanced to explore (in_progress)
$ archon workflow complete --change my-change --size small --track fast
my-change: explore marked completed
# state.yaml now: phase: explore, status: completed, track: fast, size: small
$ archon workflow advance --change my-change --to spec
my-change advanced to spec (in_progress)   # exit=0 — fast's 2nd phase, not propose
$ archon workflow advance --change my-change --to propose
blocked: propose is not part of the fast track   # exit=1 — confirms fast is truly active, not just labeled
```
This proves the full suggestion-to-execution flow end-to-end: a change bootstrapped on
`full` (the only way `explore` can be entered, since `state.yaml` must pre-exist for
anything but bootstrap) can now actually land on `fast` and have that track govern the
very next transition. This was the exact mechanism the original judge review found
broken — confirmed genuinely fixed, not just re-worded.

**Docs re-read**: `design.md`'s Data Flow section now names the real mechanism
explicitly (orchestrator branches on the preflight confirmation, calling either
`archon workflow complete --size {s} --track fast` as a single combined call, or
`--size {s}` alone) and includes a dedicated "Fix for judge finding #1" paragraph
describing the explore-only guard and why it's scoped that way (avoiding a nonsensical
mid-sequence track switch, e.g. onto `bugfix` after `design`). `specs/harness-fast-track/
spec.md`'s "Suggestion-only entry" requirement now states the same CLI call and
explicitly clarifies `advance --to {phase}`'s `--track` only bootstraps a brand-new
change and has no effect once `state.yaml` exists — consistent with the code and with
`harness-workflow/SKILL.md`'s Phase Sequence section, which gained the identical
clarification. **Verdict: FIXED, confirmed by code, tests, live binary, and docs.**

### Fix #2+3 — `harness-judge` Step 0a/0b contradiction (CRITICAL)

Re-read current `skills/harness-judge/SKILL.md` Step 0a and 0b in full (not the diff
alone). Step 0a is now the single decision point: it resolves `track` with "the SAME
back-compat rule as `trackOrder()`... absent or empty `track` resolves as `full`," then
branches three ways — non-full → `blocked` (orchestrator routing bug, stated once,
here); full+substantial/unknown → requires judge; full+small → skip-eligible. Step 0b's
`judge.enabled: true` branch no longer re-decides anything: it now reads "Step 0a
already routed a non-full track to `blocked` before computing a tier, so by this point
the resolved track is always `full`. The tier now decides..." — it purely consumes Step
0a's output. There is exactly ONE place (Step 0a) deciding non-full-track behavior now;
Step 0b's old independent "report blocked" framing was deleted, not duplicated.

Cross-checked the back-compat resolution against both references it claims to match:
`internal/workflow/workflow.go`'s `trackOrder()` — `case "", "full": return
config.PhaseOrder` — and `harness-workflow/SKILL.md`'s Step 1 Phase Sequence list —
"`full` (default; `track` absent or `full`)". Both agree: empty/absent `track` is
`full`, not a distinct unresolved state. Step 0a's wording ("Do NOT compare the raw
string to `"full"` literally — an empty string is `full`, not 'some other track'")
is explicit and matches. **Verdict: FIXED** — no remaining contradiction, and the
back-compat rule is now explicit and triple-consistent (Go, harness-workflow,
harness-judge).

### Fix #4 — `sdd-explore` Step 3c ownership (CRITICAL — duplicate-history-entry risk)

Re-read current `skills/sdd-explore/SKILL.md` Step 3c in full. It now reads: "Report
this classification to the orchestrator in your return envelope... — do NOT run `archon
workflow complete` yourself; per `skills/_shared/sdd-phase-common.md`'s executor
boundary, mutating `state.yaml` on phase completion is the orchestrator's action... The
orchestrator persists it with a SINGLE call — `archon workflow complete --change {name}
--size {small|substantial|unknown}` — combining size persistence with the ordinary
phase-completion call from `harness-workflow/SKILL.md` Step 3. There is exactly one
completion call per phase; never a second, size-only one, since `State.Complete()` has
no guard against being called twice and would append a duplicate history entry."

This correctly reassigns the actor (executor reports, orchestrator persists) and
explicitly names the single-call invariant and the reason it matters (no double-call
guard in `State.Complete()` — confirmed true by reading `internal/workflow/workflow.go`:
`Complete()` just appends a history entry and sets status, unconditionally, on every
call). The `--track` combination for the fast-track case is spelled out one layer up, in
`design.md`'s Data Flow (`archon workflow complete --size {s} --track fast (single
call...)`) and in `harness-fast-track/spec.md`'s requirement text — both already
confirmed above — so the "only completion call for the transition" invariant is
consistently described across all three documents, not just asserted in one and
silently assumed elsewhere. **Verdict: FIXED.**

### General regression check

```text
$ go build ./...
(clean, no output)
$ go vet ./...
(clean, no output)
$ go test -count=1 ./...
ok  	github.com/archon-ai/archon/cmd/archon
ok  	github.com/archon-ai/archon/internal/agent
ok  	github.com/archon-ai/archon/internal/config
ok  	github.com/archon-ai/archon/internal/initcmd
ok  	github.com/archon-ai/archon/internal/mapgen
ok  	github.com/archon-ai/archon/internal/models
ok  	github.com/archon-ai/archon/internal/opencode
ok  	github.com/archon-ai/archon/internal/route
ok  	github.com/archon-ai/archon/internal/scaffold
ok  	github.com/archon-ai/archon/internal/skillregistry
ok  	github.com/archon-ai/archon/internal/status
ok  	github.com/archon-ai/archon/internal/tui
ok  	github.com/archon-ai/archon/internal/version
ok  	github.com/archon-ai/archon/internal/workflow
ok  	github.com/archon-ai/archon/skills
```
All 15 packages pass, same set as the original verify pass — no new failures, no
newly-skipped packages.

Re-checked design coherence was not broken elsewhere by these fixes:
- `internal/config/config.go`'s `Judge` struct: confirmed zero diff since commit
  `b3bb6d1` (predates this whole change) — `git diff b3bb6d1 HEAD --
  internal/config/config.go` is empty. Untouched by both the original implementation
  and this fix commit.
- `FastTrack` is still its own named slice (`var FastTrack = []string{"explore",
  "spec", "apply", "verify", "archive"}`), not a `BugfixTrack` alias — unchanged by
  this commit.
- `trackOrder()` itself was not modified by `11c19f5` (confirmed: the commit's Go
  diff touches only `cmd/archon/main.go` and `cmd/archon/main_test.go`) — the fix is
  entirely at the CLI-flag layer, exactly matching design.md's stated scope ("the
  one point the spec allows a switch").

### Final Verdict

**PASS**

All 4 judge-identified issues are independently confirmed fixed — by source
inspection, by running the 3 new unit tests directly (not trusted from the commit
message), and by an independent empirical end-to-end run with the built binary
proving the previously-broken suggestion-to-execution flow now genuinely works
(bootstrap full → complete explore with `--size small --track fast` → `advance --to
spec` succeeds → `advance --to propose` is correctly blocked). The harness-judge
Step 0a/0b contradiction is resolved to a single decision point with an explicit,
triple-consistent back-compat rule. The sdd-explore ownership fix correctly
reassigns the `archon workflow complete` call to the orchestrator and documents the
single-call invariant consistently across `sdd-explore/SKILL.md`, `design.md`, and
`harness-fast-track/spec.md`. Full regression check (build, vet, full test suite)
is clean — same 15 packages passing as the original verify pass, no new failures.
Design coherence (`config.Judge` untouched, `FastTrack` still its own slice,
`trackOrder()` unmodified) holds.

Nothing outstanding from this re-verify. The one pre-existing WARNING (W1) from the
original pass was already resolved by the orchestrator (documented in that section
above, commit `3b2a1ca`) before this fix cycle began, and is unaffected by `11c19f5`.
No new issues found. This re-verify upgrades the change's status from the original
**PASS WITH WARNINGS** to a clean **PASS** for the specific defects this retry cycle
targeted; the change is ready to proceed to the next judge attempt.

---

## Re-Verification — 2026-10-27 (Judge Attempt 2 Fix, commit `50b249f`)

**Scope**: Independently confirm the 3 NEW issues judge attempt 2 found (introduced
by the attempt-1 fix, `11c19f5`) are genuinely fixed in `50b249f`, plus a fresh-eyes
pass hunting for a 4th round of the same contradiction class, since this exact
pattern (two sections disagreeing) has now recurred across two prior rounds.

### Issue 1 — `sdd-explore/SKILL.md` Step 3c never actually included `--track`

**Confirmed fixed.** Current Step 3c (lines ~132-139) now explicitly branches:
- `size: substantial`/`unknown`, or `small` without a confirmed `fast` suggestion →
  `archon workflow complete --change {name} --size {small|substantial|unknown}`.
- `size: small` AND confirmed `fast` suggestion →
  `archon workflow complete --change {name} --size small --track fast` — explicitly
  "`--size` and `--track` together, in the SAME call, not two calls."

This matches `design.md` line 73's documented call shape exactly
(`archon workflow complete --size {s} --track fast`). `track:fast` is now reachable
through the literal operative instruction, not just through the CLI's capability.

### Issue 2 — double-complete hazard (`--track` guard checked `state.Phase` only)

**Confirmed fixed**, by four independent checks:

1. **Source reading** (`cmd/archon/main.go` `newWorkflowCompleteCmd`, lines ~844-853):
   the `state.Status == "completed"` rejection fires immediately after `ReadState`
   and strictly before `state.Complete()`, `state.Size =`, `state.Track =`, and
   `state.Write()` — i.e. before any mutation, and unconditional on whether `--track`
   or `--size` was passed. The pre-existing phase-scoped `--track` guard
   (`trackFlag != "" && state.Phase != "explore"`) remains, now layered under the new
   check.
2. **Ran the new test**: `go test -count=1 -v ./cmd/archon/... -run
   TestWorkflowComplete_RejectsDoubleComplete` → `PASS`.
3. **Read the test body** (`cmd/archon/main_test.go:871-914`): it bootstraps a change
   to `explore`, completes it once with `--size small`, captures `state.yaml`
   (asserts exactly 2 history entries), then attempts a second `complete --track
   fast` and asserts (a) non-zero exit, (b) stderr contains "already completed", and
   (c) `state.yaml` afterward is byte-for-byte unchanged (same history length,
   `Track` still empty). This genuinely proves the claim — it's not just asserting
   an error string, it independently re-reads state off disk to prove no mutation
   occurred.
4. **Independent empirical check**, built binary in a fresh scratch dir (not reusing
   the test harness): bootstrap → `advance --to explore` → `complete --size small
   --track fast` (combined call succeeds, `state.yaml` shows `track: fast, size:
   small`) → `complete` again with no flags (rejected, "already completed",
   `state.yaml` unchanged) → `complete --size substantial --track bugfix` (also
   rejected, `state.yaml` still unchanged). All three matched expectations exactly.

### Issue 3 — `harness-judge/SKILL.md` Step 0b's unscoped "ALWAYS wins" claim

**Confirmed fixed.** Step 0b now opens with: "By this point Step 0a has already
exited (`blocked`) for any non-full track, so everything below is scoped to the
`full` track only. On the `full` track, `judge.enabled: false` ALWAYS wins..." The
"ALWAYS"/"regardless" language is no longer presented as a standalone global claim —
it is explicitly prefaced with the full-track scoping that makes it true. Read Step
0a and Step 0b together end-to-end: no remaining contradiction between the two
sections on this point.

### Fresh-eyes hunt for a 4th-round issue — FOUND ONE

Re-reading `cmd/archon/main.go`'s `newWorkflowCompleteCmd`, `sdd-explore/SKILL.md`
Step 3c, and `harness-judge/SKILL.md` Step 0 end-to-end surfaced no further
contradiction *within* those three artifacts. But the fix to issue 1 above changed
what the "correct" fast-track entry call looks like (now `--size small --track
fast`, combined) without propagating that shape to the other two documents that
describe the exact same call:

- **`skills/harness-workflow/SKILL.md` line 31**: "Mechanically, the ONLY way to
  switch an existing change onto `fast` is `archon workflow complete --change {name}
  --track fast` at the moment `explore` completes" — **omits `--size`**, and phrases
  it as "the ONLY way", i.e. exhaustively.
- **`openspec/changes/proportional-phase-gating/specs/harness-fast-track/spec.md`
  lines 85-86** (the normative spec, Requirement: "Suggestion-only entry via sizing
  classification"): "the orchestrator MUST persist the switch via `archon workflow
  complete --change {name} --track fast` at the moment `explore` completes — **the
  only CLI call** and the only phase boundary that may change an existing change's
  track" — same omission, same exhaustive phrasing.

Both of these predate `50b249f` (last touched in `11c19f5`, confirmed via `git log`)
and were never updated when issue 1's fix changed the call shape in
`sdd-explore/SKILL.md`. By contrast, `design.md` line 73 already shows the correct
combined form (`archon workflow complete --size {s} --track fast`), and
`sdd-explore/SKILL.md` now matches it — so there is a live 3-way split: 2 artifacts
say "`--track fast` alone, and that's the *only* form", 2 artifacts say "`--size` and
`--track` combined, in the same call."

This is not cosmetic. It is the exact same "two sections disagree" defect class
flagged in round 1 (original) and round-2 issue 3 (`harness-judge`), now recurring a
third time — and it has a real consequence given round 2's own issue-2 fix: because
`workflow complete` now hard-rejects any second `complete` call on an
already-completed phase, there is no way to patch in `--size` after the fact if an
orchestrator follows `harness-workflow/SKILL.md`'s or the spec's literal
documented call (`--track fast` alone) first. Following either of those two
documents as written means the size classification that justified suggesting `fast`
in the first place (`sdd-explore`'s whole reason for offering the deviation question)
is silently and permanently dropped from `state.yaml` — not recoverable, by design,
as of this very fix round. Functionally this doesn't corrupt the `fast` track's
phase sequence (judge is skipped unconditionally on `fast` regardless of tier/size),
but it breaks traceability (`archon workflow status --json`'s `size` key, and
anything reading `state.yaml` later, would show an empty size for every fast-track
change entered by the book per these two docs) and leaves a live, citable
contradiction in the normative spec itself.

**Recommended fix**: update `skills/harness-workflow/SKILL.md` line 31 and
`openspec/changes/proportional-phase-gating/specs/harness-fast-track/spec.md` lines
85-86 to show the same combined call `archon workflow complete --change {name}
--size small --track fast`, matching `design.md` and `sdd-explore/SKILL.md` exactly,
and drop the "ONLY"/"only CLI call" phrasing's implication that `--size` is absent
from it (or adjust the phrasing to say "the only call, which combines `--size` and
`--track`" to avoid re-introducing this exact ambiguity a 4th time).

### Regression Check

```
$ go build ./...
(clean, no output)
$ go vet ./...
(clean, no output)
$ go test -count=1 ./...
ok  	github.com/archon-ai/archon/cmd/archon
ok  	github.com/archon-ai/archon/internal/agent
ok  	github.com/archon-ai/archon/internal/config
ok  	github.com/archon-ai/archon/internal/initcmd
ok  	github.com/archon-ai/archon/internal/mapgen
ok  	github.com/archon-ai/archon/internal/models
ok  	github.com/archon-ai/archon/internal/opencode
ok  	github.com/archon-ai/archon/internal/route
ok  	github.com/archon-ai/archon/internal/scaffold
ok  	github.com/archon-ai/archon/internal/skillregistry
ok  	github.com/archon-ai/archon/internal/status
ok  	github.com/archon-ai/archon/internal/tui
ok  	github.com/archon-ai/archon/internal/version
ok  	github.com/archon-ai/archon/internal/workflow
ok  	github.com/archon-ai/archon/skills
```

Same 15 packages passing, no regressions from `11c19f5` or `50b249f`.

### Final Verdict

**FAIL**

All 3 issues judge attempt 2 raised are independently confirmed genuinely fixed in
`50b249f` (not rubber-stamped — each verified by source reading, and issue 2
additionally by running the new test, reading the test body for what it actually
proves, and a from-scratch empirical binary run). Full regression suite is clean.

However, the mandated fresh-eyes pass found a 4th issue of the same recurring
"two sections disagree" defect class: `skills/harness-workflow/SKILL.md` (line 31)
and the normative `specs/harness-fast-track/spec.md` (lines 85-86) still document
the fast-track entry call as `--track fast` alone (explicitly "the ONLY way" /
"the only CLI call"), while `design.md` and the just-fixed `sdd-explore/SKILL.md`
correctly show it as the combined `--size small --track fast` call. Left as-is,
this both contradicts the normative spec against its own design doc and silently
drops the size classification for any fast-track entry that follows
`harness-workflow`'s or the spec's literal instruction, with no way to repair it
after the fact given this same fix round's new hard rejection of double-completes.
Must re-apply to sync these two documents to the combined-call form before this can
pass.

---

## Verification Round 4 — 2026-10-27 (post fix round 3, commit `6a5b730`, judge retry 3/3, LAST allowed attempt)

**Scope of this pass**: adversarial, not confirmatory. Given the pattern (every prior
round missed something the next round caught), this pass re-derives every claim from
source rather than trusting the fix-round-3 commit message's "exhaustive sweep found
none" assertion.

### 1. Side-by-side read of the 5 canonical mechanism documents

Read in full, word-for-word, not skimmed:
- `skills/sdd-explore/SKILL.md` Step 3c (lines 101-139)
- `skills/harness-workflow/SKILL.md` "Phase Sequence" `fast` bullet (lines 25-36) and
  Step 1 (lines 107-113)
- `skills/harness-judge/SKILL.md` Step 0 end-to-end (lines 37-106)
- `openspec/changes/proportional-phase-gating/design.md` Data Flow (lines 64-96)
- `openspec/changes/proportional-phase-gating/specs/harness-fast-track/spec.md`
  "Suggestion-only entry" requirement (lines 79-112)

**Result: no drift between any pair.** Specifically cross-checked and confirmed
identical on every point that previously drifted:
- The combined-call form `archon workflow complete --change {name} --size small
  --track fast` (both flags, one call) is stated identically in all 5 (sdd-explore
  L138-139, harness-workflow L31, design.md L73, spec.md L86, and spec.md's own
  scenario at L92-101 exercises it in substance).
- `--track` on `advance` bootstraps only, ignored once `state.yaml` exists, and is
  rejected by `complete` on every phase except `explore` — stated identically in
  harness-workflow L33-36, design.md L89-96/126-128, spec.md L88-90.
- `judge.enabled: false` "ALWAYS wins" is explicitly scoped to "On the full track"
  in harness-judge L75 (attempt 2's finding #3 fix held).
- Fast's 5-phase sequence (`explore → spec → apply → verify → archive`, no judge)
  and its full-Gherkin (not bugfix-scoped) spec phase are stated identically in
  harness-workflow L25-28/L36, design.md L1-9 header comment + "Decision: fast gets
  its own named slice" (L41-52), and spec.md Requirements 1-2.
- Literal grep of every occurrence of `--track fast` with the `--` prefix across the
  whole repo (`grep -rn -- "--track fast"`) confirms every current, operative hit is
  immediately preceded by `--size small ` / `--size {s} ` on the same line in
  design.md:73, spec.md:86, and harness-workflow/SKILL.md:31 — no bare-form survivor
  anywhere in the 5 documents, re-confirming (not just trusting) fix round 3's claim.

No drift found among the 5 documents themselves.

### 2. Build / vet / full test suite

```text
$ go build ./...        → clean, exit 0
$ go vet ./...           → clean, exit 0
$ go test -count=1 ./... → all 15 packages ok, 0 failures
```

### 3. Empirical re-run of the full flow with the built binary (`go install ./cmd/archon`)

Fresh scratch dir, real binary, no mocking:

1. `archon workflow advance --change demo-change --to explore` → bootstrapped, `allowed`.
2. `archon workflow complete --change demo-change --size small --track fast` (single
   combined call) → `demo-change: explore marked completed`. `state.yaml` afterward:
   `phase: explore, status: completed, track: fast, size: small` — confirmed by
   reading the file directly.
3. `archon workflow advance --change demo-change --to spec` → `{"allowed":true,...}`.
4. `archon workflow advance --change demo-change --to propose` → blocked:
   `"propose is not part of the fast track"` — exact match to spec.md's scenario
   "Propose is not reachable in fast track".
5. Double-complete guard: completed `spec` once (succeeded), then called `complete`
   again twice — once with no flags, once with `--size substantial --track full`.
   Both rejected: `Error: spec is already completed — run archon workflow advance to
   start the next phase, not complete again`. `md5sum`/`diff` of `state.yaml` before
   and after both rejected attempts: **identical, zero mutation**, confirmed for both
   the flagless and the flagged repro.
6. `--track` scoping: advanced to `apply`, then ran `advance --to apply --track
   bugfix` on the already-existing change → succeeded as a no-op resume
   (`resuming:true`) with `track:` unchanged at `fast` in `state.yaml` (confirms
   "`--track` on `advance` ... ignored once `state.yaml` exists"). Then ran `complete
   --change demo-change --track full` while on `apply` (non-explore phase) →
   rejected: `"--track is only accepted while completing the explore phase (current
   phase: apply)"`, `state.yaml`'s `track:` field unchanged.
7. Invalid-flag rejection: `complete --size huge` and `complete --track stacked` both
   rejected before any write (`diff` against a pre-call snapshot: no change).
   `advance --to explore --track bogus` on a brand-new change: rejected
   (`"unknown track: bogus ..."`), and no `state.yaml` was ever created for that
   change — confirming the bootstrap path is equally guarded.

Every scenario the spec and design documents claim, reproduced exactly as claimed,
with no deviation.

### 4. Fresh skeptical read of `newWorkflowCompleteCmd` and `CheckTransition`/`trackOrder`

Read both in full as if never seen. `newWorkflowCompleteCmd` (`cmd/archon/main.go`
805-878): validates `--size`/`--track` against fixed allow-lists before touching
`state.yaml` (817-826); rejects a second `complete` on an already-`completed` phase
by checking **`state.Status`**, not `state.Phase` (844-848 — this is attempt 2's
fix, re-confirmed still correct: a `Phase`-only check would have let a second
`complete --track ...` land on the completed `explore` phase before advancing, which
is exactly the double-complete hazard fix round 2 closed); rejects `--track` on any
phase other than `explore` (849-853), checked *after* the completed-guard so a
double-complete attempt with `--track` hits the right error message first (confirmed
empirically in run 5 above: the error returned was the already-completed one, not
the track-scoping one, because `spec` was already completed — order of checks is
correct for that case; separately confirmed the track-scoping check in isolation at
`apply`, an `in_progress`, not-yet-completed phase). No ordering bug found: the two
guards are independent preconditions (one on `Status`, one on `Phase`) and their
checked order doesn't change which calls are accepted — only which message a
rejected call receives.

`CheckTransition`/`trackOrder` (`internal/workflow/workflow.go` 41-231): `trackOrder`
only recognizes `"", "full", "bugfix", "fast"`, anything else errors — never silently
coerces (36-54 in the same file's exported constant, `MaxJudgeRetries` unrelated).
`CheckTransition`'s bootstrap branch (`currentPhase == ""`) only allows
`requestedIdx == 0`; its steady-state `switch` (206-230) is a clean, exhaustive
partition over `{==, +1, >+1, <=current}` with no gap — re-verified by hand against
the Decision Gates table in `harness-workflow/SKILL.md` line by line, no mismatch.

### 5. Repo-wide stale-reference sweep (redone independently, not trusting the prior commit's claim)

Reran the sweep with a different method than fix round 3 presumably used (grep for
the literal `--` prefixed form specifically, across the whole repo, no file-type
filter) — see section 1's grep result. No bare `--track fast` survivor in any
currently-operative instruction file. The only hits without `--size` on the same
line are either plain-English Gherkin prose ("suggests track fast" /
"sets track fast") in `spec.md`/`harness-workflow/spec.md`, which are not CLI
invocations, or `verify-report.md`'s own historical narrative of **past** findings
(expected — a changelog, not current guidance).

**However, this sweep surfaced a genuine NEW (previously uncaught) drift, outside
the 5 canonical documents the brief named, but squarely an "other stale reference to
the fast-track mechanism"**:

#### Finding: this repo's own rendered `CLAUDE.md` and `AGENTS.md` were never regenerated from `internal/initcmd/templates.go` after this change

`internal/initcmd/templates.go` is the template source `RenderClaudeMD`/
`RenderAgentsMD` render into a project's root `CLAUDE.md`/`AGENTS.md` (via `archon
init`, confirmed at `internal/initcmd/init.go:297-324`, `writeTemplate` — a full
overwrite, not a merge, so re-rendering is safe and requires no hand-reconciliation).
This repo is itself an archon-managed project (it has its own `SESSION_STATUS.md`,
`openspec/changes/`, and both `CLAUDE.md` and `AGENTS.md` at its root) and both files
were last touched at commit `56eb262` — which predates `6dc3af3` (this change's
original `feat` commit) and all three fix rounds (`11c19f5`, `50b249f`, `6a5b730`).

Concretely, comparing `templates.go`'s current source against the two rendered
files:

- `templates.go:57-63` (`orchestratorSections`, shared by both renderers): the
  Deviation rule has a 4th bullet, `(d) sdd-explore classifies the change size as
  small (suggest track: fast via one scoped question — never auto-selected)`.
  **`CLAUDE.md:54-56` and `AGENTS.md`'s equivalent Deviation rule stop at bullet
  (c)** — bullet (d) is entirely absent from both rendered files.
- `templates.go:156` / `templates.go:172` (`orchestratorRulesClaude` Rule 6 /
  `orchestratorRulesOpencode` Rule 6): "After verify, invoke harness-judge, which
  gates the `full` track's judge phase on its own computed per-change risk tier
  (track + size), not an unconditional pass — an explicit `judge.enabled: false`
  always wins over the tier; `bugfix` and `fast` never run judge". **`CLAUDE.md:134`
  and `AGENTS.md:110` both still read the old, pre-tiering text: "6. After verify,
  invoke harness-judge"** — no mention of the computed tier, `size`, or `fast` at
  all.
- Confirmed this is specific to this change, not a pre-existing gap: `track: bugfix`
  content (from the earlier `f60dcb9` bugfix-track feature) **is** present in both
  rendered files (`CLAUDE.md:91`, `AGENTS.md:66`), proving the render-and-commit step
  was done for that earlier feature. It was simply never done for this one, across
  the original apply and all three subsequent fix rounds — `tasks.md`'s tasks 3.5/
  3.6 and `apply-progress.md`'s matching entries scope the deliverable to
  `internal/initcmd/templates.go` only; regenerating this repo's own root files was
  never a task, so no fix round had a checklist item that would have caught it.

**Why this matters, not just as a documentation nit**: `specs/harness-fast-track/
spec.md`'s "Suggestion-only entry" requirement depends on the orchestrator's
Deviation rule actually containing bullet (d) in order to ever surface the one
scoped question. In this repo, right now, an orchestrator session operates on
`CLAUDE.md` (confirmed: this is literally the file injected as this session's
"project instructions... checked into the codebase" and stated to "OVERRIDE any
default behavior"). Its Deviation rule has no bullet (d) and its Rule 6 says nothing
about tiering — so a real orchestrator session in this repo, today, would never
suggest `track: fast` no matter what `sdd-explore` classifies, and would treat
`harness-judge` as an unconditional gate rather than a tier-computed one. The feature
this entire change exists to add is implemented correctly in the Go binary and the
skills, but is unreachable through this repo's own live orchestrator prompt.

**Fix required**: re-render `CLAUDE.md` and `AGENTS.md` from the current
`templates.go` (a full overwrite via the existing `writeTemplate`/`RenderClaudeMD`/
`RenderAgentsMD` path, or equivalent) so both pick up Deviation-rule bullet (d) and
the reworded Rule 6. No other content in either file needs to change — both are
fully template-generated with no hand-added sections layered on top (verified by
reading each file in full against `templates.go`'s concatenated constants).

### 6. `tasks.md` coherence

23/23 checked, matching `apply-progress.md`'s own status line. The 3 fix rounds since
original apply were not added as new tasks — confirmed this is expected/correct
(judge-driven corrections, not new scope) per the brief, and consistent with how
`harness-judge`'s retry loop is documented to work (re-apply with feedback, not a
tasks.md edit). No incoherence found. (Separately, task 3.5/3.6's scoping to
`templates.go` only, rather than to this repo's own rendered files, is the root cause
of section 5's finding — noted there, not a `tasks.md` defect in itself: the tasks
accurately describe what they scoped, they just didn't scope widely enough.)

### Verdict

**FAIL**

All mechanism-level claims across the 5 canonical documents are now mutually
consistent, the Go implementation matches every one of them exactly, the full test
suite is clean, and seven independent empirical CLI scenarios (combined-call entry,
blocked propose, double-complete guard with confirmed zero mutation under two flag
variants, post-bootstrap `--track` no-op on `advance`, phase-scoped `--track`
rejection on `complete`, and invalid-flag rejection with zero mutation including at
bootstrap) all reproduced exactly as the documents claim. No new defect was found in
`internal/workflow/workflow.go`, `cmd/archon/main.go`, or any of the 5 canonical
documents themselves.

However, this pass found a 5th-round defect of a new kind (not "two prose documents
disagree with each other," like rounds 1-3, but "the live, deployed orchestrator
prompt for this very project was never updated to match the documents it's supposed
to summarize"): this repo's own root `CLAUDE.md` and `AGENTS.md` are stale relative
to `internal/initcmd/templates.go`, missing both the Deviation rule's `fast`-suggestion
bullet and the tier-computed `harness-judge` wording for Rule 6. This makes the
`track: fast` entry path the whole change exists to add practically unreachable from
this repo's own orchestrator session, despite being correctly implemented everywhere
else. Must re-render (or hand-edit to match) `CLAUDE.md` and `AGENTS.md` from the
current `templates.go` before this can pass.

---

## Round 6 Re-Verify (2026-10-27) — Targeted Re-Check of Commit `108322a`

Scope: a narrow re-verify of `108322a` ("docs(sdd): sync this repo's own
CLAUDE.md/AGENTS.md with templates.go"), the fix for round 5's finding above, plus a
mandated full-file drift sweep (not limited to the two sections the commit targeted)
and a sanity check for other committed, templates.go-rendered output elsewhere in the
repo.

### 1. Targeted fix verified byte-for-byte

Confirmed programmatically, not by eye: wrote a temporary Go test
(`internal/initcmd/zzdump_test.go`, deleted after use — `git status --short` showed
no leftover files) that calls `RenderClaudeMD`/`RenderAgentsMD` directly from the
`initcmd` package with `TemplateData` matching this repo's actual installed state
(`SkillCount: 23`, the 9 Claude phase models and 8 opencode phase models exactly as
listed in the committed files), wrote the output to the scratchpad, and ran
`diff -u` against the real root `CLAUDE.md`/`AGENTS.md`. Both diffs show **zero
difference** for the Deviation-rule bullet (d) and Rule 6 text in both files — the
`§sdd-explore§ classifies the change size as §small§...` bullet and the
`...gates the §full§ track's judge phase on its own computed per-change risk
tier...` Rule 6 wording are present verbatim (post `§`→backtick substitution) in
both `CLAUDE.md` and `AGENTS.md`. Round 5's specific finding is fully closed.

Also confirmed via `git show 108322a -- CLAUDE.md AGENTS.md`: the commit made
exactly the two surgical hand-edits it claims (deviation bullet + Rule 6 in both
files) and touched nothing else — consistent with its commit message's claim of a
"surgical" sync rather than a full regeneration.

### 2. Full-file drift sweep — found 3 additional, PRE-EXISTING drift spots unrelated to this change

Per the brief's instruction to cross-check every section `templates.go` defines (not
just the two this commit targeted), the same diff-the-real-render approach surfaced
three more drift spots. All three predate `proportional-phase-gating` entirely (confirmed
via `git log -p --follow` on each file) and are **not** caused by, and were never in
scope for, this change's apply/fix rounds. Reporting them here because the brief
explicitly asked for this sweep and because they are the same meta-failure-mode
(committed copy vs. generator going stale) that this whole verify thread exists to
catch:

- **`CLAUDE.md` "## Phase Order" section has ~11 lines of hand-added content with no
  counterpart in `templates.go`**: the "In the single-PR flow the archive commit is
  staged BEFORE the PR is opened..." / "In the Feature Branch Chain flow..." /
  "Stacked-to-Main is unsupported..." paragraphs. `git log` traces these to commits
  `ff2ace5`, `fac3e5f`, `3f0f24d`, `b38ce88`, `bfcf203` (all predating this change,
  explicitly described as hand-edits, e.g. `549955e` "feat(docs): hand-edit CLAUDE.md
  and AGENTS.md..."). `internal/initcmd/templates.go` has **no** "Terminal Phase
  Ordering" / "Stacked-to-Main Archive Convergence" text anywhere
  (`grep -rn "Terminal Phase Ordering\|Stacked-to-Main Archive Convergence"
  internal/initcmd/` returns nothing). `AGENTS.md` does **not** have this extra
  content in its own Phase Order section — it matches `templates.go` exactly there.
  So this is CLAUDE.md-only hand-added content that was never backported into the
  shared generator.
- **`CLAUDE.md`'s Session Status archive bullet also carries extra wording** not in
  `templates.go` ("...as part of the staged archive commit, before the PR is opened
  (single-PR flow) or on the tracker branch before the tracker PR merges to `main`
  (Feature Branch Chain flow)..." vs. the template's plain "...into the archived
  change folder alongside the feature artifacts..."). Same root cause as above.
  `AGENTS.md` matches `templates.go` verbatim on this bullet.
- **`AGENTS.md` is missing the entire "## Concise Chat Output" section** (14 lines)
  that `templates.go`'s shared `orchestratorSections` constant defines for *both*
  renderers, and that `CLAUDE.md` does have. Traced to `d096c77` (Jul 2026, "feat:
  add concise-output skill..."), whose own commit message says: "CLAUDE.md: surgical
  additive insertion of the same pointer (no regen)" — `AGENTS.md` was never touched.
- **`AGENTS.md`'s "## Phase Models" section header text contradicts
  `templates.go`'s `phaseModelsOpencode` constant**: the committed file reads
  "Advisory: when delegating an SDD phase, request the model below... This is a
  preference, not a hard gate..." while `templates.go` (confirmed consistent with
  its own test `TestTemplates_AgentsPhaseModelsPointsAtOpencodeJSON`) says "...The
  binding lives in `opencode.json` under `agent.archon-<phase>.model`" — an opposite
  framing (advisory vs. binding). Traced to `d714180`, the original commit that
  added `AGENTS.md`, which predates the "binding" wording being written into
  `templates.go`.

None of these four lines of drift touch the Deviation rule or Rule 6 text this
change's fix rounds were scoped to — confirmed by diff: the only hunks the
`CLAUDE.md`/`AGENTS.md` vs. rendered-template diff produces are the four items
above, nothing else. No new drift was introduced by `108322a` itself; `108322a`'s
own two edits match the template exactly.

**Disposition**: flagging as a warning, not a blocker for this change. These four
items are pre-existing technical debt from earlier, unrelated features
(chained-pr/archive-before-PR hand-edits, the concise-output skill rollout, and
`AGENTS.md`'s original authoring) — none were introduced or worsened by
`proportional-phase-gating` or by `108322a`. Recommend a separate follow-up change to
either (a) backport the Feature-Branch-Chain Phase-Order content and the hard-gate
Phase-Models wording into `templates.go` so `archon init` on new projects doesn't
ship an incomplete/inconsistent file, and add the missing Concise Chat Output section
and matching Phase-Models wording to this repo's own `AGENTS.md`, or (b) accept the
CLAUDE.md-only additions as intentionally out-of-template and document that
exception explicitly in `templates.go`'s comments so future sweeps don't re-flag it
every round.

### 3. Build / vet / full test suite

```
go build ./...      → clean, no output
go vet ./...         → clean, no output
go test -count=1 ./... → ok, all 15 packages (cmd/archon, internal/agent,
                          internal/config, internal/initcmd, internal/mapgen,
                          internal/models, internal/opencode, internal/route,
                          internal/scaffold, internal/skillregistry,
                          internal/status, internal/tui, internal/version,
                          internal/workflow, skills)
```

As expected — `108322a` touched no `.go` files, so this is a formality. Confirmed
clean nonetheless.

### 4. Sanity check — any other committed, templates.go-rendered output elsewhere in the repo?

Checked:
- `find . -iname "CLAUDE.md" -o -iname "AGENTS.md"` → exactly one of each, both at
  the repo root. No duplicate/secondary rendered copies anywhere else (e.g. no
  per-package or per-doc copies).
- `.opencode/skills/*` (`git ls-files .opencode/` confirms these ARE git-tracked):
  all entries are **symlinks** (`lrwxrwxrwx`) pointing to
  `/home/alexcasdev/.config/opencode/skills/<name>` — a different machine/user's
  home directory, so they resolve as broken links in this environment. These are
  machine-local skill-registry links materialized by `archon init`/`update`
  (`internal/initcmd/refresh.go`, `internal/initcmd/update.go`), not rendered text
  content, and not derived from `templates.go` at all (`templates.go` only renders
  `CLAUDE.md`/`AGENTS.md`). This is a separate, pre-existing symlink-portability
  quirk, out of scope for this check, noted only for completeness — it is not a
  "stale rendering of `templates.go`" issue.
- `skills/` (the actual skill source directory): grepped for text unique to the
  orchestrator-prompt sections (`"Deviation rule"`, `"Vague Request Guard"`,
  `"Human Review Gate"`, `"Leader Persona"`) — the only hit outside
  `CLAUDE.md`/`AGENTS.md`/`openspec/changes/**` is `skills/concise-output/SKILL.md`,
  which quotes the Human Review Gate string as a cross-reference (expected — it's
  the skill `templates.go`'s pointer line names), not a duplicated full rendering.

**Conclusion**: no other committed, generated rendering of `templates.go` exists in
this repo besides the root `CLAUDE.md`/`AGENTS.md` themselves.

### Round 6 Verdict

**PASS WITH WARNINGS**

The specific defect this re-verify was scoped to — `108322a`'s sync of the
Deviation-rule bullet (d) and Rule 6 wording into `CLAUDE.md`/`AGENTS.md` — is
confirmed correct and complete by direct programmatic diff against the actual
`templates.go` renderer, not a visual skim. Build, vet, and the full test suite are
clean. No other committed `templates.go` rendering exists elsewhere in the repo.

The warning: the mandated full-file sweep found four lines of **pre-existing**
drift between `templates.go` and the committed `CLAUDE.md`/`AGENTS.md`, unrelated
to and not introduced by `proportional-phase-gating` or `108322a` — (1) CLAUDE.md's
Phase Order section carries ~11 lines of Feature-Branch-Chain content with no
`templates.go` counterpart, (2) CLAUDE.md's Session Status archive bullet has
matching extra wording, (3) AGENTS.md is missing the entire "Concise Chat Output"
section, and (4) AGENTS.md's Phase Models section text contradicts `templates.go`'s
current "hard gate" wording with stale "advisory" framing. These do not block this
change's archive — they are orthogonal, older technical debt — but should be tracked
as a separate follow-up so the next unrelated feature's sync doesn't get blamed for
them.
