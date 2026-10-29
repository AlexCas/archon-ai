Feature: harness-judge delegates the dual review to archon-judge
  harness-judge runs the dual adversarial review by delegating to the
  frontmatter-pinned archon-judge subagent instead of running judgment-day inline,
  while preserving the enabled gate, mutation/playwright gates, the 3-retry loop,
  and the feedback contract.

  Background:
    Given an initialized claude project with judge enabled

  # Requirement: Judgment-Day Wrapping
  Scenario: Judgment-day passes
    Given sdd-verify completed successfully for change "my-feature"
    When harness-judge runs
    Then the dual review is delegated to the archon-judge subagent
    And if both judges approve, the verdict is "pass"
    And state.yaml advances to judge phase with status "completed"

  Scenario: Judgment-day finds issues
    Given the archon-judge dual review returns one or more issues
    When harness-judge processes the verdict
    Then the verdict is "fail"
    And all issues are collected into structured feedback

  @edge
  Scenario: Dual review runs under the pinned judge model
    Given archon-judge.md frontmatter pins "model" to "claude-opus-4-8"
    When harness-judge delegates the dual review
    Then judgment-day executes under the archon-judge subagent's pinned model
    And not under the orchestrator's model

  # Requirement: Computed risk tier gates judge (full track only)
  @edge
  Scenario: Explicit judge.enabled false always skips, regardless of tier
    Given .archon/config.yaml has judge.enabled set to false
    And the computed risk tier would otherwise require judge
    When harness-judge runs Step 0
    Then the archon-judge subagent is not invoked
    And the gate is skipped exactly as before

  @edge
  Scenario: Low-risk computed tier skips or lightens judge
    Given a full-track change and judge.enabled is not explicitly false
    And the computed risk tier is skip-eligible (low risk)
    When harness-judge runs Step 0
    Then the archon-judge subagent is not invoked, or runs a lightened review
    And archive is reachable without a blocking full review

  @edge
  Scenario: Re-apply loop and retry cap are unchanged
    Given the single review returns verdict "fail" with structured feedback
    When harness-judge processes the failure
    Then sdd-apply is auto-invoked with the feedback, then sdd-verify, then a re-judge
    And after 3 consecutive failures harness-judge returns "blocked" with max_retries_exceeded true

  # Requirement: Computed Per-Change Risk Tier
  @happy
  Scenario: Substantial full-track change always requires judge
    Given a full-track change classified substantial by sdd-explore
    When harness-judge computes the risk tier at Step 0
    Then the tier requires the full single judge

  @edge
  Scenario: Small full-track change computes a skip-eligible tier
    Given a full-track change classified small by sdd-explore
    And judge.enabled is not explicitly false
    When harness-judge computes the risk tier at Step 0
    Then the tier is skip-eligible
