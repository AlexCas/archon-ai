Feature: Computed risk tier gates the judge phase
  Step 0 computes a per-change risk tier from track and sdd-explore's size
  classification instead of reading a flat static boolean, while an explicit
  judge.enabled: false always wins.

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
