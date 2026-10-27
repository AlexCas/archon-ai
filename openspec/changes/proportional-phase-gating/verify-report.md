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
