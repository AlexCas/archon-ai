# harness-bugfix-track Specification

<!-- [[harness-bugfix-track]] · [proposal](../../proposal.md) · [exploration](../../exploration.md) -->

## Purpose

Defines an abbreviated SDD track for bug fixes. A change may declare `track: bugfix`
in its `state.yaml`, selecting a 5-phase sequence (explore → spec → apply → verify →
archive) with no judge phase and a minimal three-section scoped spec, instead of the
full 9-phase track. The `track` field is back-compatible: absent means `full`.

Scenarios for each requirement are in `harness-bugfix-track.feature` alongside this file.

## Requirements

### Requirement: Track field on state.yaml

`state.yaml` MUST support an optional `track` field with values `full` or `bugfix`.
When the field is absent, the change MUST behave as `track: full`. Setting
`track: bugfix` MUST select the bugfix phase sequence and its gates. No other value
MUST be accepted; an unknown value MUST be reported as an error rather than silently
treated as `full`.

#### Scenario: Absent track defaults to full

```gherkin
@happy
Scenario: Absent track defaults to full
  Given a change whose state.yaml has no track field
  When the harness resolves the change track
  Then the track is treated as full
  And the full 9-phase sequence applies
```

#### Scenario: Explicit bugfix track is honored

```gherkin
@happy
Scenario: Explicit bugfix track is honored
  Given a change whose state.yaml has track set to bugfix
  When the harness resolves the change track
  Then the bugfix 5-phase sequence applies
```

#### Scenario: Unknown track value is rejected

```gherkin
@error
Scenario: Unknown track value is rejected
  Given a change whose state.yaml has track set to "hotfix"
  When the harness resolves the change track
  Then an error is reported naming the supported values full and bugfix
  And the change is not silently treated as full
```

### Requirement: Bugfix phase sequence

The bugfix track MUST enforce the phase order `explore → spec → apply → verify →
archive`. The phases `propose`, `design`, `tasks`, and `judge` MUST NOT appear in the
bugfix sequence and MUST NOT be reachable within a `track: bugfix` change. Within the
track, phases MUST NOT be skipped: each phase MUST complete before the next is allowed.

#### Scenario: Bugfix advances explore to spec

```gherkin
@happy
Scenario: Bugfix advances explore to spec
  Given a bugfix change with explore completed
  When the orchestrator requests the spec phase
  Then the transition is allowed
  And the next phase after spec is apply
```

#### Scenario: Propose is not reachable in bugfix track

```gherkin
@error
Scenario: Propose is not reachable in bugfix track
  Given a bugfix change with explore completed
  When the orchestrator requests the propose phase
  Then the transition is blocked
  And the response indicates propose is not part of the bugfix track
```

#### Scenario: Spec cannot be skipped in bugfix track

```gherkin
@edge
Scenario: Spec cannot be skipped in bugfix track
  Given a bugfix change with explore completed
  When the orchestrator requests the apply phase before spec completes
  Then the transition is blocked
  And the response names spec as the required next phase
```

### Requirement: Judge absent in bugfix track

The bugfix track MUST NOT run a judge phase. After `verify` completes, the next and
final phase MUST be `archive`. The judge phase MUST be omitted regardless of the value
of `judge.enabled` in `.archon/config.yaml` — that flag governs the full track only.

#### Scenario: Verify transitions straight to archive

```gherkin
@happy
Scenario: Verify transitions straight to archive
  Given a bugfix change with verify completed
  When the orchestrator requests the next phase
  Then the allowed next phase is archive
  And no judge phase is required or run
```

#### Scenario: judge.enabled does not add a judge to the bugfix track

```gherkin
@edge
Scenario: judge.enabled does not add a judge to the bugfix track
  Given a bugfix change and .archon/config.yaml with judge.enabled true
  When the bugfix change reaches verify completed
  Then archive is the next phase
  And the judge phase is still omitted
```

### Requirement: Manual verification in bugfix track

Because the bugfix track omits the automated judge gate, `verify` MUST be the final
correctness gate before archive, and the Human Review Gate MUST fire once — after the
scoped spec. Verification of a bugfix MAY be performed manually, but the verify phase
MUST still run and MUST record its outcome in `state.yaml` before archive is allowed.

#### Scenario: Human Review Gate fires once after the scoped spec

```gherkin
@happy
Scenario: Human Review Gate fires once after the scoped spec
  Given a bugfix change producing a scoped spec
  When the spec phase completes
  Then the Human Review Gate fires for the scoped spec
  And it does not fire again before apply
```

#### Scenario: Archive is blocked until verify is recorded

```gherkin
@error
Scenario: Archive is blocked until verify is recorded
  Given a bugfix change with apply completed but verify not recorded
  When the orchestrator requests archive
  Then the transition is blocked
  And the response names verify as the required next phase
```

### Requirement: Scoped spec format contract

In the bugfix track, `spec` MUST produce a single `spec.md` in the change folder with
exactly three sections and NO Gherkin suite: `## Bug` (observed vs expected behavior,
1–3 sentences), `## Fix Criteria` (a flat checklist of 2–5 verifiable acceptance
bullets), and `## Non-Regression` (1–2 bullets naming existing behavior that must not
break). The scoped spec MUST NOT perform capability or requirement decomposition and
MUST NOT emit `.feature` files.

#### Scenario: Scoped spec has exactly the three required sections

```gherkin
@happy
Scenario: Scoped spec has exactly the three required sections
  Given a bugfix change entering the spec phase
  When the scoped spec is written
  Then spec.md contains a "## Bug" section, a "## Fix Criteria" section, and a "## Non-Regression" section
  And it contains no other requirement sections
```

#### Scenario: Fix Criteria is a flat verifiable checklist

```gherkin
@happy
Scenario: Fix Criteria is a flat verifiable checklist
  Given a bugfix scoped spec being written
  When the Fix Criteria section is authored
  Then it lists between 2 and 5 checklist bullets
  And each bullet states a verifiable condition
```

#### Scenario: Scoped spec emits no Gherkin feature file

```gherkin
@edge
Scenario: Scoped spec emits no Gherkin feature file
  Given a bugfix change in the spec phase
  When the scoped spec is written
  Then no .feature file is created in the change folder
  And spec.md contains no Gherkin Scenario blocks
```
