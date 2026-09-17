# Delta for harness-judge

<!-- [[harness-judge]] · [proposal](../../proposal.md) · [exploration](../../exploration.md) -->

This delta replaces the dual-adversarial review in the SDD path with a single focused
judge. The `judgment-day` skill is explicitly untouched and remains a standalone opt-in.
The enabled gate, the mutation/playwright gates, the 3-retry re-apply loop, and the
feedback contract are preserved.

## MODIFIED Requirements

### Requirement: Single Focused Judge in the SDD Path

In the SDD path, the meta-skill MUST run exactly ONE focused review by DELEGATING to the
`archon-judge` subagent (whose frontmatter `model:` is the binding hard gate). The single
judge MUST evaluate spec compliance, design coherence, and code quality, and MUST return
a single verdict (`pass`/`fail`) with any issues. The meta-skill MUST NOT run a parallel
dual/blind review, MUST NOT synthesize across two judges, and MUST NOT run any review
inline on the orchestrator's model. The captured verdict drives the rest of the gate.
(Previously: `harness-judge` delegated a DUAL adversarial review — two blind judges plus
synthesis — to `archon-judge`, which invoked `judgment-day`.)

#### Scenario: Single judge passes

```gherkin
Scenario: Single judge passes
  Given sdd-verify completed successfully for change "my-feature"
  When harness-judge runs
  Then a single focused review is delegated to the archon-judge subagent
  And if the review approves, the verdict is "pass"
  And state.yaml advances to judge phase with status "completed"
```

#### Scenario: Single judge finds issues

```gherkin
Scenario: Single judge finds issues
  Given the archon-judge single review returns one or more issues
  When harness-judge processes the verdict
  Then the verdict is "fail"
  And all issues are collected into structured feedback
```

#### Scenario: Review runs under the pinned judge model, not a dual review

```gherkin
@edge
Scenario: Review runs under the pinned judge model, not a dual review
  Given archon-judge.md frontmatter pins "model" to "claude-opus-4-8"
  When harness-judge delegates the review
  Then the single focused review executes under the archon-judge subagent's pinned model
  And not under the orchestrator's model
  And no second judge or blind-review synthesis is run
```

### Requirement: Preserved gates and loop under the single judge

The single-judge model MUST NOT change any other harness-judge behavior. The
`judge.enabled` gate, the mutation-testing gate, the Playwright gate, the maximum
3-retry auto-re-apply loop, and the structured feedback output contract MUST all behave
exactly as they did under the prior delegation.
(Previously: these gates were described as preserved "under delegation" of the dual
review; the same gates and loop now wrap the single focused judge.)

#### Scenario: Disabled judge skips the review

```gherkin
@edge
Scenario: Disabled judge skips the review
  Given .archon/config.yaml has judge disabled
  When harness-judge runs
  Then the archon-judge subagent is not invoked
  And the gate is skipped exactly as before
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

### Requirement: judgment-day Remains Standalone and Untouched

The `judgment-day` skill MUST remain unchanged by the single-judge model and MUST NOT be
invoked by the SDD path. It MUST remain a standalone, opt-in skill activated only when the
user explicitly requests it (e.g. "juzgar", "dual review", "adversarial review"). The SDD
`judge` phase MUST NOT trigger `judgment-day` automatically.

#### Scenario: SDD judge does not invoke judgment-day

```gherkin
@happy
Scenario: SDD judge does not invoke judgment-day
  Given a full-track change reaching the SDD judge phase
  When harness-judge runs the single focused review
  Then judgment-day is not invoked
  And only the single archon-judge review is performed
```

#### Scenario: judgment-day still runs on explicit request

```gherkin
@happy
Scenario: judgment-day still runs on explicit request
  Given the user explicitly asks for a dual/adversarial review ("juzgar")
  When the request is processed
  Then judgment-day runs its standalone dual review unchanged
  And this is independent of the SDD judge phase
```
