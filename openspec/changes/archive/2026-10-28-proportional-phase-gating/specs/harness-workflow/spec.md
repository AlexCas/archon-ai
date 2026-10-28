# Delta for harness-workflow

<!-- [[harness-workflow]] · [proposal](../../proposal.md) · [exploration](../../exploration.md) -->

This delta lets the full-track judge gate be computed per change instead of an
unconditional static pass, and adds `fast` as a third `track` value governed by
[[harness-fast-track]]. The Decision Gates table, retry persistence, and preflight
prose are otherwise unchanged.

## MODIFIED Requirements

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

## ADDED Requirements

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
