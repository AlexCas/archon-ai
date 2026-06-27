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

  # Requirement: Preserved gates and loop under delegation
  @edge
  Scenario: Disabled judge skips the delegated review
    Given .archon/config.yaml has judge disabled
    When harness-judge runs
    Then the archon-judge subagent is not invoked
    And the gate is skipped exactly as before

  @edge
  Scenario: Re-apply loop and retry cap are unchanged
    Given the delegated dual review returns verdict "fail" with structured feedback
    When harness-judge processes the failure
    Then sdd-apply is auto-invoked with the feedback, then sdd-verify, then a re-judge
    And after 3 consecutive failures harness-judge returns "blocked" with max_retries_exceeded true
