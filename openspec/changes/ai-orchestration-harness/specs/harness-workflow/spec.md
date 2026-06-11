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
