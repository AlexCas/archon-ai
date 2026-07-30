# harness-workflow Specification

## Purpose

The `harness-workflow` meta-skill enforces the SDD phase sequence (explore → propose → spec → design → tasks → apply → verify → judge → archive) by reading change state and blocking invalid transitions.

## Requirements

### Requirement: Phase State Machine

The meta-skill MUST enforce a linear phase progression with exactly one allowed next phase per state.

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

### Requirement: State Persistence

The meta-skill MUST read and write change state from `openspec/changes/{name}/state.yaml`.

#### Scenario: State read on invocation

- GIVEN `openspec/changes/my-feature/state.yaml` contains `phase: proposed, status: completed`
- WHEN `harness-workflow` is invoked for `my-feature`
- THEN it reads the current phase and status before enforcing any transition

#### Scenario: State updated on transition

- GIVEN transition from `proposed` to `spec` is allowed
- WHEN the transition is approved
- THEN `state.yaml` is updated to `phase: spec, status: in_progress`
- AND a timestamp is recorded

### Requirement: Workflow Reporting

The meta-skill MUST report current phase and allowed transitions on request.

#### Scenario: Report current state

- GIVEN change `my-feature` is in `tasks` phase with status `completed`
- WHEN the orchestrator requests state report
- THEN `harness-workflow` returns `current: tasks, status: completed, next: apply`

### Requirement: Phase Skipping Prevention

The meta-skill MUST NOT allow skipping mandatory phases (propose, spec, design, tasks, apply, verify).

#### Scenario: Attempt to skip from propose to apply

- GIVEN the change state is `proposed`
- WHEN the orchestrator requests `apply`
- THEN `harness-workflow` returns `blocked`
- AND reports all missing phases: `spec, design, tasks`

### Requirement: Automatic map.md Regen on Phase Transition

After every successful phase transition, the harness-workflow transition path MUST
trigger `archon map` to regenerate `openspec/map.md`. This MUST happen after the
`state.yaml` update so the new phase and status are reflected in the generated index.
The regen MUST be transparent to the orchestrator: a regen failure SHOULD be reported
as a warning but MUST NOT block the phase transition from being recorded.

#### Scenario: map.md is regenerated after a successful phase transition

```gherkin
@happy
Scenario: map.md is regenerated after a successful phase transition
  Given a change my-feature transitioning from spec to design
  When harness-workflow approves the transition and updates state.yaml
  Then archon map is invoked to regenerate openspec/map.md
  And map.md reflects my-feature with phase=design and status=in_progress
```

#### Scenario: Regen failure does not block the transition

```gherkin
@error
Scenario: Regen failure does not block the transition
  Given archon map exits non-zero during a phase transition regen
  When harness-workflow detects the regen failure
  Then the phase transition is still recorded in state.yaml
  And a warning about the regen failure is surfaced to the orchestrator
  And the transition is not rolled back
```

#### Scenario: Regen runs after state.yaml is written

```gherkin
@edge
Scenario: Regen runs after state.yaml is written
  Given a phase transition from tasks to apply
  When harness-workflow processes the transition
  Then state.yaml is updated before archon map is invoked
  And the generated map.md shows the updated phase and status
```

### Requirement: Terminal Phase Ordering (Single-PR Flow)

The harness-workflow MUST enforce archive-before-PR ordering in the single-PR flow:
after `judge` passes, the archive step MUST run (and its commit MUST be staged on the
change branch) BEFORE the PR is opened. This ensures the archive commit travels inside
the change's own PR with no trailing archive-only commit.

Judge gating is UNCHANGED: archive MUST NOT run until judge passes.

The archive-internal operation order MUST be preserved:

1. Spec merge (delta spec merged into `openspec/specs/{domain}/spec.md`)
2. Folder move (`openspec/changes/{name}/` → `openspec/changes/archive/YYYY-MM-DD-{name}/`)
3. `SESSION_STATUS.md` move into the archived folder (within the same commit)
4. `archon map --backfill` + `archon map --check`

All four steps MUST be staged into a single archive commit. PR-open is EXTERNAL to
this sequence — the PR is opened only after the archive commit is on the branch.

`SESSION_STATUS.md` MUST remain root-resident from session start through judge
completion. It MUST be moved into the archived change folder only when the archive
commit is being staged, satisfying the `[[session-status-contract]]` invariant
("root-resident during work, moved at archive"). The move happening before PR-open
(instead of after) does NOT violate this invariant.

#### Scenario: Judge passes, archive runs, then PR is opened

- GIVEN the change has reached `judge` phase with status `completed` (judge passed)
- AND `SESSION_STATUS.md` exists at the repository root
- WHEN the orchestrator begins the terminal sequence
- THEN archive operations run (spec merge, folder move, SESSION_STATUS.md move, archon map)
- AND all archive operations are staged into a single archive commit on the change branch
- AND ONLY AFTER the archive commit is staged, the PR is opened
- AND the archive commit is part of the change's PR (no separate archive-only commit after PR creation)

#### Scenario: Archive is not attempted before judge passes

- GIVEN the change is in `verify` phase with status `completed`
- WHEN the orchestrator attempts to start archive
- THEN `harness-workflow` returns `blocked`
- AND the response indicates `judge` must complete before `archive` can run
- AND no archive operations are performed

#### Scenario: SESSION_STATUS.md stays root-resident through judge

- GIVEN the change has passed `verify` and is entering `judge`
- AND `SESSION_STATUS.md` exists at the repository root
- WHEN judge runs and completes (pass or fail)
- THEN `SESSION_STATUS.md` remains at the repository root
- AND it is NOT moved until the archive commit is being staged

#### Scenario: SESSION_STATUS.md moves in the archive commit (before PR-open)

- GIVEN judge has passed and archive is running
- WHEN the archive commit is being staged
- THEN `SESSION_STATUS.md` is moved from the repository root into
  `openspec/changes/archive/YYYY-MM-DD-{change-name}/SESSION_STATUS.md`
- AND this move is part of the same archive commit as the spec merge and folder move
- AND the PR is opened only AFTER this commit is staged
- AND `SESSION_STATUS.md` is no longer present at the repository root after the commit

#### Scenario: Archive-internal ordering is preserved

- GIVEN judge has passed and archive is executing
- WHEN the four archive sub-operations run
- THEN spec merge completes before folder move begins
- AND folder move completes before `archon map --backfill` is invoked
- AND `archon map --check` runs after `--backfill` completes
- AND PR-open is not invoked until all four sub-operations and the archive commit are done

#### Scenario: Chained-PR flow is out of scope (deferred to slice 2)

- GIVEN a change using the chained-PR flow (multiple stacked PRs)
- WHEN the terminal phase sequence is evaluated
- THEN the archive-before-PR ordering rule from this requirement DOES NOT apply
- AND the harness-workflow treats chained-PR archive ownership as undefined / slice-2 pending
- AND no blocking or enforcement of archive position is applied to chained-PR flows
