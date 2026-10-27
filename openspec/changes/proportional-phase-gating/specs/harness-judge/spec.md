# Delta for harness-judge

<!-- [[harness-judge]] · [proposal](../../proposal.md) · [exploration](../../exploration.md) -->

This delta replaces the static `judge.enabled` boolean read at Step 0 with a computed
per-change risk tier, while preserving the explicit `judge.enabled: false` override so
existing configs keep working unchanged. The single-judge delegation model, the
mutation/Playwright gates, and the 3-retry loop are untouched.

## MODIFIED Requirements

### Requirement: Computed risk tier gates judge (full track only)

Step 0 MUST read [[harness-judge]]'s computed per-change risk tier to decide whether
the `full`-track judge phase runs, lightens, or is skipped — NOT a static boolean
alone. This tiering applies to the `full` track only: [[harness-fast-track]] and
`[[harness-bugfix-track]]` never run judge, regardless of tier. If `judge.enabled` is
literally `false` in `.archon/config.yaml`, that explicit setting ALWAYS skips judge
regardless of the computed tier — the computed tier only matters when `judge.enabled`
is not explicitly `false`. The mutation-testing gate, the Playwright gate, the maximum
3-retry auto-re-apply loop, and the structured feedback output contract MUST all
behave exactly as before.
(Previously: "Preserved gates and loop under the single judge" — `judge.enabled` was a
flat static boolean with no computed tiering.)

#### Scenario: Explicit judge.enabled false always skips, regardless of tier

```gherkin
@edge
Scenario: Explicit judge.enabled false always skips, regardless of tier
  Given .archon/config.yaml has judge.enabled set to false
  And the computed risk tier would otherwise require judge
  When harness-judge runs Step 0
  Then the archon-judge subagent is not invoked
  And the gate is skipped exactly as before
```

#### Scenario: Low-risk computed tier skips or lightens judge

```gherkin
@edge
Scenario: Low-risk computed tier skips or lightens judge
  Given a full-track change and judge.enabled is not explicitly false
  And the computed risk tier is skip-eligible (low risk)
  When harness-judge runs Step 0
  Then the archon-judge subagent is not invoked, or runs a lightened review
  And archive is reachable without a blocking full review
```

#### Scenario: Re-apply loop and retry cap are unchanged

```gherkin
@edge
Scenario: Re-apply loop and retry cap are unchanged
  Given the single review returns verdict "fail" with structured feedback
  When harness-judge processes the failure
  Then sdd-apply is auto-invoked with the feedback, then sdd-verify, then a re-judge
  And after 3 consecutive failures harness-judge returns "blocked" with max_retries_exceeded true
```

## ADDED Requirements

### Requirement: Computed Per-Change Risk Tier

`harness-judge` Step 0 MUST compute the risk tier from the change's `track` and
`sdd-explore`'s size classification (`small`/`substantial`/`unknown`). A `full`-track
change classified `substantial` or `unknown` MUST compute a tier that requires the
full single judge. A `full`-track change classified `small` MAY compute a
skip-eligible tier. The computed tier MUST NOT override an explicit
`judge.enabled: false`, which always wins per the requirement above.

#### Scenario: Substantial full-track change always requires judge

```gherkin
@happy
Scenario: Substantial full-track change always requires judge
  Given a full-track change classified substantial by sdd-explore
  When harness-judge computes the risk tier at Step 0
  Then the tier requires the full single judge
```

#### Scenario: Small full-track change computes a skip-eligible tier

```gherkin
@edge
Scenario: Small full-track change computes a skip-eligible tier
  Given a full-track change classified small by sdd-explore
  And judge.enabled is not explicitly false
  When harness-judge computes the risk tier at Step 0
  Then the tier is skip-eligible
```
