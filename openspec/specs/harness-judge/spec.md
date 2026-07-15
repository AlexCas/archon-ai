# Delta for harness-judge

This delta routes the dual review through the frontmatter-pinned `archon-judge`
subagent instead of running `judgment-day` inline on the orchestrator's model. All
other harness-judge behavior — the enabled gate, mutation/playwright gates, the
3-retry re-apply loop, and the feedback contract — is preserved unchanged.

## MODIFIED Requirements

### Requirement: Judgment-Day Wrapping

The meta-skill MUST run the dual adversarial review by DELEGATING to the
`archon-judge` subagent (whose frontmatter `model:` is the binding hard gate), which
invokes `judgment-day` and reports its verdict. The meta-skill MUST NOT run
`judgment-day` inline on the orchestrator's model. The captured verdict drives the
rest of the gate exactly as before.
(Previously: `harness-judge` invoked `judgment-day` inline, inheriting the orchestrator's model.)

#### Scenario: Judgment-day passes

```gherkin
Scenario: Judgment-day passes
  Given sdd-verify completed successfully for change "my-feature"
  When harness-judge runs
  Then the dual review is delegated to the archon-judge subagent
  And if both judges approve, the verdict is "pass"
  And state.yaml advances to judge phase with status "completed"
```

#### Scenario: Judgment-day finds issues

```gherkin
Scenario: Judgment-day finds issues
  Given the archon-judge dual review returns one or more issues
  When harness-judge processes the verdict
  Then the verdict is "fail"
  And all issues are collected into structured feedback
```

#### Scenario: Dual review runs under the pinned judge model

```gherkin
@edge
Scenario: Dual review runs under the pinned judge model
  Given archon-judge.md frontmatter pins "model" to "claude-opus-4-8"
  When harness-judge delegates the dual review
  Then judgment-day executes under the archon-judge subagent's pinned model
  And not under the orchestrator's model
```

## ADDED Requirements

### Requirement: Preserved gates and loop under delegation

Routing the dual review through `archon-judge` MUST NOT change any other
harness-judge behavior. The `judge.enabled` gate, the mutation-testing gate, the
Playwright gate, the maximum 3-retry auto-re-apply loop, and the structured feedback
output contract MUST all behave exactly as they did when `judgment-day` ran inline.

#### Scenario: Disabled judge skips the delegated review

```gherkin
@edge
Scenario: Disabled judge skips the delegated review
  Given .archon/config.yaml has judge disabled
  When harness-judge runs
  Then the archon-judge subagent is not invoked
  And the gate is skipped exactly as before
```

#### Scenario: Re-apply loop and retry cap are unchanged

```gherkin
@edge
Scenario: Re-apply loop and retry cap are unchanged
  Given the delegated dual review returns verdict "fail" with structured feedback
  When harness-judge processes the failure
  Then sdd-apply is auto-invoked with the feedback, then sdd-verify, then a re-judge
  And after 3 consecutive failures harness-judge returns "blocked" with max_retries_exceeded true
```
