# Delta for harness-workflow

<!-- [[harness-workflow]] · [proposal](../../proposal.md) · [exploration](../../exploration.md) -->

This delta makes the phase state machine track-aware (full vs bugfix), removes the
engram/none artifact-store scope branches (artifacts are ALWAYS OpenSpec), reflects the
single-judge model in the full-track terminal ordering, and adds the preflight-as-prose
rule (silent standard default, deviation-triggered proposal only).

A later delta (code-backed-workflow) ports the Decision Gates table this spec describes
into `internal/workflow` (Go, with its own test suite) and exposes it as `archon workflow
status|advance|complete|retry`. Before that delta, every rule here was enforced only by
an orchestrating LLM choosing to read and follow this prose correctly — with no code-level
check and no automated test of the actual enforcement.

## MODIFIED Requirements

### Requirement: Phase State Machine

The meta-skill MUST enforce a linear phase progression with exactly one allowed next
phase per state, selecting the phase sequence from the change's `track`:

- `full` (default; `track` absent or `full`): `explore → propose → spec → design →
  tasks → apply → verify → judge → archive`.
- `bugfix`: `explore → spec → apply → verify → archive` (governed by
  `[[harness-bugfix-track]]`; `propose`, `design`, `tasks`, and `judge` are not part
  of the sequence).

The meta-skill MUST read `track` before building the transition table and MUST NOT
allow a phase that is not part of the selected track.
(Previously: a single hard-coded 9-phase sequence with no track concept.)

#### Scenario: Valid transition is allowed

- GIVEN the current change state is `proposed`
- WHEN the orchestrator requests `spec` phase
- THEN `harness-workflow` returns `allowed` and records the new state as `specifying`

#### Scenario: Invalid transition is blocked

- GIVEN the current change state is `proposed`
- WHEN the orchestrator requests `tasks` phase (skipping spec and design)
- THEN `harness-workflow` returns `blocked`
- AND the response includes the required next phase: `spec`

#### Scenario: Phase in-progress is idempotent

- GIVEN the current change state is `designing` (in progress)
- WHEN the orchestrator requests `design` again
- THEN `harness-workflow` returns `allowed` with status `resuming`

#### Scenario: Track selects the phase sequence

```gherkin
@happy
Scenario: Track selects the phase sequence
  Given a change whose state.yaml has track set to bugfix
  When harness-workflow builds the transition table
  Then the sequence is explore, spec, apply, verify, archive
  And propose, design, tasks, and judge are not reachable
```

### Requirement: Phase Skipping Prevention

The meta-skill MUST NOT allow skipping the mandatory phases of the change's selected
track. For the `full` track these are propose, spec, design, tasks, apply, verify. For
the `bugfix` track these are spec, apply, verify (per `[[harness-bugfix-track]]`). A
phase that does not belong to the selected track MUST NOT be treated as a skippable or
requestable phase within that change.
(Previously: named a single fixed set of mandatory phases with no track awareness.)

#### Scenario: Attempt to skip from propose to apply

- GIVEN the change state is `proposed`
- WHEN the orchestrator requests `apply`
- THEN `harness-workflow` returns `blocked`
- AND reports all missing phases: `spec, design, tasks`

#### Scenario: Bugfix track skip is blocked

```gherkin
@error
Scenario: Bugfix track skip is blocked
  Given a bugfix change with explore completed
  When the orchestrator requests apply before spec completes
  Then harness-workflow returns blocked
  And the response names spec as the required next phase
```

### Requirement: Fast Track Selection in Terminal Ordering

`track` MUST additionally accept `fast` as a third value, alongside `full` and
`bugfix`. The `fast` phase sequence and its gates are governed entirely by
[[harness-fast-track]]. `fast` MUST only be entered via `sdd-explore`'s sizing
suggestion and a confirmed preflight deviation question — the orchestrator MUST NOT
silently auto-select it.

#### Scenario: Fast track is resolved alongside full and bugfix

```gherkin
@happy
Scenario: Fast track is resolved alongside full and bugfix
  Given a change whose state.yaml has track set to fast
  When harness-workflow builds the transition table
  Then the sequence is governed by harness-fast-track
  And propose, design, tasks, and judge are not reachable
```

#### Scenario: Fast is never silently auto-selected

```gherkin
@error
Scenario: Fast is never silently auto-selected
  Given sdd-explore classifies a change's size as small
  When the orchestrator prepares the track decision
  Then track fast is only set after the user confirms the scoped suggestion
  And harness-workflow never sets track fast without that confirmation
```

### Requirement: Stacked-to-Main Archive Convergence

When the orchestrator is selecting a chain strategy during `sdd-tasks`, it MUST NOT
select pure Stacked-to-Main. It MUST select (or silently convert to) **Feature Branch
Chain** instead. This decision is **up-front and decisive**: made at chain-strategy
selection in `sdd-tasks`, before any child PR is opened. Because artifacts are ALWAYS
OpenSpec, archive-before-PR is ALWAYS in effect; there is no artifact-store mode under
which this requirement is inapplicable.
(Previously: the requirement was scoped to "when archive-before-PR is in effect" and
carried an explicit exemption for `engram`/`none` artifact-store modes; those modes are
removed.)

**Rationale:** Stacked-to-Main ships each slice independently to `main`. By the time an
archive commit is needed there is no single un-merged owning ref; the archive-before-PR
invariant ("archive commit inside the change's PR, not separate") cannot be satisfied.
Feature Branch Chain supplies the required tracker branch.

**Up-front convergence is mandatory:** A late Stacked-to-Main → FBC conversion
(attempted after one or more slices have already merged to `main`) is NOT the sanctioned
path. It would strand already-merged slices on `main` without an owning tracker ref,
violating the archive-before-PR invariant retroactively. The orchestrator MUST converge
at selection time.

**Unsupported combination:** Pure independent-shipping Stacked-to-Main is explicitly
**unsupported**. If the user requests Stacked-to-Main, the orchestrator MUST inform them
that this combination is unsupported and MUST select Feature Branch Chain instead. No
partial `main` merges are permitted before an owning tracker ref exists.

**After convergence:** The change is now Feature Branch Chain. The "Terminal Phase
Ordering (Feature Branch Chain)" requirement governs the archive commit — staged on the
tracker branch, after the integrated judge passes, before the tracker PR merges to
`main`. Zero additional archive mechanics are introduced by this requirement.

#### Scenario: Stacked-to-Main requested — orchestrator converges to FBC at sdd-tasks

- GIVEN the orchestrator is selecting a chain strategy during `sdd-tasks`
- AND the user's preference or default would be Stacked-to-Main
- WHEN the orchestrator executes `sdd-tasks` strategy selection
- THEN the orchestrator selects Feature Branch Chain instead of Stacked-to-Main
- AND the orchestrator notifies the user that pure Stacked-to-Main is unsupported and FBC is selected
- AND a tracker/integration branch is created (or planned) before any child PR is opened
- AND the archive step will be governed by the "Terminal Phase Ordering (Feature Branch Chain)" requirement

#### Scenario: Pure Stacked-to-Main is unsupported — no partial main merges before owning ref

- GIVEN the orchestrator attempted to proceed with pure Stacked-to-Main (no tracker branch)
- WHEN any child PR would merge to `main` without a tracker branch in place
- THEN the `sdd-apply` backstop returns `blocked`
- AND the response states that pure Stacked-to-Main is an unsupported combination
- AND no child PR is merged to `main` until the strategy is corrected to Feature Branch Chain
- AND `harness-workflow` directs the orchestrator to converge to FBC before opening child PRs

#### Scenario: After convergence, FBC archive rule governs — no new mechanics

- GIVEN the orchestrator has converged from Stacked-to-Main to Feature Branch Chain at sdd-tasks
- AND the change is now proceeding as a Feature Branch Chain flow
- AND the integrated judge has passed on the tracker branch
- WHEN the terminal phase sequence runs
- THEN the archive step follows the "Terminal Phase Ordering (Feature Branch Chain)" requirement verbatim
- AND no additional archive mechanics are introduced by the Stacked-to-Main convergence
- AND the archive commit is staged on the tracker branch before the tracker PR merges to `main`

#### Scenario: Late Stacked-to-FBC conversion is out of scope — stranded slices hazard

- GIVEN one or more slices have already merged to `main` under Stacked-to-Main
- WHEN the orchestrator attempts a late conversion to Feature Branch Chain
- THEN this conversion path is NOT the sanctioned approach
- AND `harness-workflow` treats this state as an unresolved hazard (stranded slices on `main` without an owning tracker ref)
- AND the orchestrator MUST NOT attempt archive under these conditions without explicit human resolution
- AND this scenario exists to document the hazard, not to define a recovery procedure

## ADDED Requirements

### Requirement: Single Judge in the Full-Track Terminal Ordering

In the `full` track, whether judge runs is gated by [[harness-judge]]'s Step 0
computed per-change risk tier, not an unconditional pass. When the tier requires
review, exactly ONE judge phase MUST run after `verify` and before `archive`,
performing a single focused review (per `[[harness-judge]]`); the terminal ordering
MUST NOT invoke a second judge or a parallel dual review as part of the SDD path.
When the tier is skip-eligible (low risk), judge MAY be skipped or lightened and
`archive` MAY proceed directly from `verify`. If `judge.enabled` is literally `false`
in `.archon/config.yaml`, judge is ALWAYS skipped regardless of the computed tier.
Whenever judge does run, `archive` MUST NOT run until it passes; both the single-PR
and Feature Branch Chain terminal orderings treat a required judge's pass as the
archive gate.
(Previously: judge ran unconditionally on every full-track change, with no
risk-tiering or skip path.)

#### Scenario: One judge gates archive when the tier requires review

```gherkin
@happy
Scenario: One judge gates archive when the tier requires review
  Given a full-track change with verify completed
  And the computed risk tier requires judge
  When the terminal sequence runs
  Then exactly one judge phase runs before archive
  And archive is blocked until that single judge passes
```

#### Scenario: No second judge is added to the SDD path

```gherkin
@edge
Scenario: No second judge is added to the SDD path
  Given a full-track change reaching the judge phase
  When harness-workflow orders the terminal phases
  Then it invokes a single focused judge, not a parallel dual review
  And no additional judge phase is inserted before archive
```

#### Scenario: Low-risk tier skips judge and archive proceeds from verify

```gherkin
@edge
Scenario: Low-risk tier skips judge and archive proceeds from verify
  Given a full-track change with verify completed
  And the computed risk tier is skip-eligible
  When the terminal sequence runs
  Then judge is skipped or lightened
  And archive is reachable directly from verify
```

#### Scenario: Explicit judge.enabled false always skips, regardless of tier

```gherkin
@edge
Scenario: Explicit judge.enabled false always skips, regardless of tier
  Given .archon/config.yaml has judge.enabled set to false
  And the computed risk tier would otherwise require judge
  When the terminal sequence runs
  Then judge is skipped
  And archive proceeds directly from verify
```

### Requirement: Code-Backed Transition Validation and State Mutation

The Decision Gates table (Requirements "Phase State Machine" and "Phase Skipping
Prevention" above) MUST be implemented in code (`internal/workflow.CheckTransition`),
not re-derived by the orchestrating model from this prose on every transition. The
meta-skill MUST call `archon workflow advance --change {name} --to {phase}` to validate
and perform a transition, `archon workflow complete --change {name}` to mark the current
phase completed, and MUST NOT hand-edit `state.yaml` to perform either. For claude
projects, a `PreToolUse` hook (`internal/initcmd`'s `state-yaml-guard.sh`) MUST block any
direct `Edit`/`Write` tool call targeting `openspec/changes/*/state.yaml`, so this is not
only a prose convention.

`state.yaml` gains a `retry_count` field (default 0) persisting `[[harness-judge]]`'s
auto-re-apply attempt count, managed exclusively via `archon workflow retry
increment|reset`, so the 3-retry cap survives a context compaction or a closed-and-
reopened session instead of living only in the orchestrating model's in-conversation
count.

`archon workflow status --change {name}` MUST render the "Current phase" line and the
"Phase History" checklist that `[[session-status-contract]]` specifies for
`SESSION_STATUS.md`, directly from `state.yaml`. The orchestrator MUST paste that
rendered block verbatim rather than hand-typing those two sections, since `state.yaml`
and `SESSION_STATUS.md` are two independent sources of truth for the same phase state
that can otherwise drift apart.

#### Scenario: Advance delegates to the code-level check

```gherkin
@happy
Scenario: Advance delegates to the code-level check
  Given a change whose state.yaml has phase "spec" and status "completed"
  When the orchestrator runs "archon workflow advance --change my-feature --to design"
  Then the command reports allowed
  And state.yaml is updated atomically with phase "design", status "in_progress", and a new history entry
  And the orchestrator does not reimplement the Decision Gates table inline
```

#### Scenario: Direct state.yaml edits are blocked for claude projects

```gherkin
@edge
Scenario: Direct state.yaml edits are blocked for claude projects
  Given a claude project with archon's hooks installed
  When any agent attempts to Edit or Write openspec/changes/{name}/state.yaml directly
  Then the PreToolUse hook blocks the call
  And the hook's message names "archon workflow advance/complete/retry" as the sanctioned path
```

#### Scenario: Retry count persists across a session restart

```gherkin
@edge
Scenario: Retry count persists across a session restart
  Given a change's state.yaml has retry_count 2 from a prior judge failure
  And the session that incremented it was closed and a new one started
  When the new session runs "archon workflow retry increment --change my-feature"
  Then the reported retry_count is 3, not 1
  And a subsequent increment attempt reports max_retries_exceeded: true
```

#### Scenario: SESSION_STATUS.md's phase block is generated, not hand-typed

```gherkin
@happy
Scenario: SESSION_STATUS.md's phase block is generated, not hand-typed
  Given a change's state.yaml has phase "apply" and status "in_progress"
  When the orchestrator runs "archon workflow status --change my-feature"
  Then the output includes "**Current phase**: apply (in_progress)"
  And a "## Phase History" checklist matching state.yaml's history exactly
  And the orchestrator pastes this block into SESSION_STATUS.md instead of typing it by hand
```

### Requirement: Preflight is a Standard Default Expressed as Prose

The SDD session preflight MUST be a fixed standard default expressed as prose in the
orchestrator files (`CLAUDE.md`/`AGENTS.md`) and their source template — NOT a Go config
struct and NOT a per-session multi-group questionnaire. The standard default is:
Ritmo=interactivo, Artefactos=OpenSpec, PRs=ask-but-default-chained, Revisión=800 lines.
The default path MUST be applied SILENTLY (no preflight ceremony). The orchestrator MUST
surface a scoped one-line decision ONLY on a genuine deviation: (a) the estimate exceeds
the 800-line budget, (b) the change is trivially small (suggest a single PR), or (c) the
user explicitly asks to adjust. Absent a deviation, no preflight question is asked.

#### Scenario: Standard default is applied silently

```gherkin
@happy
Scenario: Standard default is applied silently
  Given a fresh SDD session with no deviation signal
  When the orchestrator begins an SDD phase
  Then the standard preflight default is applied without asking a preflight question
  And no multi-group preflight questionnaire is shown
```

#### Scenario: Large estimate triggers a scoped deviation question

```gherkin
@edge
Scenario: Large estimate triggers a scoped deviation question
  Given a change whose estimate exceeds the 800-line review budget
  When the orchestrator prepares the PR strategy
  Then it surfaces a single scoped question about adjusting the budget or slicing
  And it does not re-run the full preflight ceremony
```

#### Scenario: Trivially small change suggests a single PR

```gherkin
@edge
Scenario: Trivially small change suggests a single PR
  Given a trivially small change well under the review budget
  When the orchestrator prepares the PR strategy
  Then it may suggest a single PR instead of a chain
  And it asks at most one scoped question
```
