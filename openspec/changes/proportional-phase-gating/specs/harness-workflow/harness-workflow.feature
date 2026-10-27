Feature: Proportional judge gating and the fast track
  Full-track judge gating is computed per change (risk tier) instead of
  unconditional, and track resolution gains a third value, fast, governed by
  harness-fast-track.

  @happy
  Scenario: One judge gates archive when the tier requires review
    Given a full-track change with verify completed
    And the computed risk tier requires judge
    When the terminal sequence runs
    Then exactly one judge phase runs before archive
    And archive is blocked until that single judge passes

  @edge
  Scenario: No second judge is added to the SDD path
    Given a full-track change reaching the judge phase
    When harness-workflow orders the terminal phases
    Then it invokes a single focused judge, not a parallel dual review
    And no additional judge phase is inserted before archive

  @edge
  Scenario: Low-risk tier skips judge and archive proceeds from verify
    Given a full-track change with verify completed
    And the computed risk tier is skip-eligible
    When the terminal sequence runs
    Then judge is skipped or lightened
    And archive is reachable directly from verify

  @edge
  Scenario: Explicit judge.enabled false always skips, regardless of tier
    Given .archon/config.yaml has judge.enabled set to false
    And the computed risk tier would otherwise require judge
    When the terminal sequence runs
    Then judge is skipped
    And archive proceeds directly from verify

  @happy
  Scenario: Fast track is resolved alongside full and bugfix
    Given a change whose state.yaml has track set to fast
    When harness-workflow builds the transition table
    Then the sequence is governed by harness-fast-track
    And propose, design, tasks, and judge are not reachable

  @error
  Scenario: Fast is never silently auto-selected
    Given sdd-explore classifies a change's size as small
    When the orchestrator prepares the track decision
    Then track fast is only set after the user confirms the scoped suggestion
    And harness-workflow never sets track fast without that confirmation
