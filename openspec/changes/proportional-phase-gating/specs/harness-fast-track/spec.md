# harness-fast-track Specification

<!-- [[harness-fast-track]] · [proposal](../../proposal.md) · [exploration](../../exploration.md) -->

## Purpose

Abbreviated SDD track for small, well-understood feature work: `track: fast` selects
explore → spec → apply → verify → archive, no judge — sibling to
[[harness-bugfix-track]], but for features, so `spec` stays a full Gherkin spec.
Entry is suggestion-only: `sdd-explore` classifies size and the orchestrator asks one
scoped question; `fast` is never auto-selected silently. Scenarios live in
`harness-fast-track.feature` beside this file.

## Requirements

### Requirement: Track field supports fast

`state.yaml`'s `track` field MUST accept `fast` as a third value, alongside `full` and
`bugfix` (resolution per [[harness-workflow]]). `track: fast` MUST select the fast
sequence below; its `spec` phase MUST produce a full Gherkin spec, never the bugfix
3-section scoped format.

#### Scenario: Explicit fast track is honored

```gherkin
@happy
Scenario: Explicit fast track is honored
  Given a change whose state.yaml has track set to fast
  When the harness resolves the change track
  Then the fast 5-phase sequence applies
  And its spec phase is a full Gherkin spec, not the scoped bugfix format
```

### Requirement: Fast phase sequence

The fast track MUST enforce `explore → spec → apply → verify → archive`. The phases
`propose`, `design`, `tasks`, and `judge` MUST NOT appear and MUST NOT be reachable
within a `track: fast` change. Phases MUST NOT be skipped.

#### Scenario: Fast advances explore to spec

```gherkin
@happy
Scenario: Fast advances explore to spec
  Given a fast change with explore completed
  When the orchestrator requests the spec phase
  Then the transition is allowed
```

#### Scenario: Propose is not reachable in fast track

```gherkin
@error
Scenario: Propose is not reachable in fast track
  Given a fast change with explore completed
  When the orchestrator requests the propose phase
  Then the transition is blocked
  And the response indicates propose is not part of the fast track
```

### Requirement: Judge absent in fast track

The fast track MUST NOT run a judge phase, regardless of `judge.enabled` or
[[harness-judge]]'s computed risk tier — risk-tiering governs the `full` track only.
After `verify` completes, the next and final phase MUST be `archive`.

#### Scenario: Verify transitions straight to archive regardless of tier

```gherkin
@happy
Scenario: Verify transitions straight to archive regardless of tier
  Given a fast change with verify completed
  And a computed risk tier that would require judge on the full track
  When the orchestrator requests the next phase
  Then the allowed next phase is archive
  And no judge phase is required or run
```

### Requirement: Suggestion-only entry via sizing classification

`sdd-explore` MUST classify a change's size as `small`, `substantial`, or `unknown`. For
`small`, the orchestrator MAY suggest `track: fast` via the existing preflight
deviation-question pattern (one scoped question) — the orchestrator MUST NOT silently
auto-select `fast`. On `unknown`, the track MUST default to `full`.

#### Scenario: Small classification suggests fast via one scoped question

```gherkin
@happy
Scenario: Small classification suggests fast via one scoped question
  Given sdd-explore classifies the change size as small
  When the orchestrator prepares the track decision
  Then it surfaces exactly one scoped question suggesting track fast
  And it does not select fast without the user's confirmation
```

#### Scenario: Unknown classification defaults to full

```gherkin
@edge
Scenario: Unknown classification defaults to full
  Given sdd-explore cannot classify the change size (unknown)
  When the orchestrator prepares the track decision
  Then the track defaults to full
  And fast is not suggested
```

### Requirement: Human Review Gate preserved

The fast track MUST NOT skip the Human Review Gate for phases it retains. The gate MUST
fire after `spec` (fast's only gated artifact phase), exactly as it would under `full`.

#### Scenario: Human Review Gate fires once after spec

```gherkin
@happy
Scenario: Human Review Gate fires once after spec
  Given a fast change producing a full spec
  When the spec phase completes
  Then the Human Review Gate fires for the spec
  And it does not fire again before apply
```
