Feature: harness-workflow triggers automatic map.md regen
  After every successful phase transition, harness-workflow invokes archon map
  so map.md is always current without manual intervention.

  @happy
  Scenario: map.md is regenerated after a successful phase transition
    Given a change my-feature transitioning from spec to design
    When harness-workflow approves the transition and updates state.yaml
    Then archon map is invoked to regenerate openspec/map.md
    And map.md reflects my-feature with phase=design and status=in_progress

  @error
  Scenario: Regen failure does not block the transition
    Given archon map exits non-zero during a phase transition regen
    When harness-workflow detects the regen failure
    Then the phase transition is still recorded in state.yaml
    And a warning about the regen failure is surfaced to the orchestrator
    And the transition is not rolled back

  @edge
  Scenario: Regen runs after state.yaml is written
    Given a phase transition from tasks to apply
    When harness-workflow processes the transition
    Then state.yaml is updated before archon map is invoked
    And the generated map.md shows the updated phase and status

  # Requirement: Single Judge in the Full-Track Terminal Ordering (tier-computed)
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

  # Requirement: Fast Track Selection in Terminal Ordering
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
