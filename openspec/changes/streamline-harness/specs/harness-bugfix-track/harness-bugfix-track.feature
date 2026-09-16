Feature: Harness bugfix track
  A change can declare an abbreviated bugfix track that runs
  explore -> spec (scoped) -> apply -> verify -> archive, with no judge phase.

  # Track field on state.yaml

  @happy
  Scenario: Absent track defaults to full
    Given a change whose state.yaml has no track field
    When the harness resolves the change track
    Then the track is treated as full
    And the full 9-phase sequence applies

  @happy
  Scenario: Explicit bugfix track is honored
    Given a change whose state.yaml has track set to bugfix
    When the harness resolves the change track
    Then the bugfix 5-phase sequence applies

  @error
  Scenario: Unknown track value is rejected
    Given a change whose state.yaml has track set to "hotfix"
    When the harness resolves the change track
    Then an error is reported naming the supported values full and bugfix
    And the change is not silently treated as full

  # Bugfix phase sequence

  @happy
  Scenario: Bugfix advances explore to spec
    Given a bugfix change with explore completed
    When the orchestrator requests the spec phase
    Then the transition is allowed
    And the next phase after spec is apply

  @error
  Scenario: Propose is not reachable in bugfix track
    Given a bugfix change with explore completed
    When the orchestrator requests the propose phase
    Then the transition is blocked
    And the response indicates propose is not part of the bugfix track

  @edge
  Scenario: Spec cannot be skipped in bugfix track
    Given a bugfix change with explore completed
    When the orchestrator requests the apply phase before spec completes
    Then the transition is blocked
    And the response names spec as the required next phase

  # Judge absent in bugfix track

  @happy
  Scenario: Verify transitions straight to archive
    Given a bugfix change with verify completed
    When the orchestrator requests the next phase
    Then the allowed next phase is archive
    And no judge phase is required or run

  @edge
  Scenario: judge.enabled does not add a judge to the bugfix track
    Given a bugfix change and .archon/config.yaml with judge.enabled true
    When the bugfix change reaches verify completed
    Then archive is the next phase
    And the judge phase is still omitted

  # Manual verification in bugfix track

  @happy
  Scenario: Human Review Gate fires once after the scoped spec
    Given a bugfix change producing a scoped spec
    When the spec phase completes
    Then the Human Review Gate fires for the scoped spec
    And it does not fire again before apply

  @error
  Scenario: Archive is blocked until verify is recorded
    Given a bugfix change with apply completed but verify not recorded
    When the orchestrator requests archive
    Then the transition is blocked
    And the response names verify as the required next phase

  # Scoped spec format contract

  @happy
  Scenario: Scoped spec has exactly the three required sections
    Given a bugfix change entering the spec phase
    When the scoped spec is written
    Then spec.md contains a "## Bug" section, a "## Fix Criteria" section, and a "## Non-Regression" section
    And it contains no other requirement sections

  @happy
  Scenario: Fix Criteria is a flat verifiable checklist
    Given a bugfix scoped spec being written
    When the Fix Criteria section is authored
    Then it lists between 2 and 5 checklist bullets
    And each bullet states a verifiable condition

  @edge
  Scenario: Scoped spec emits no Gherkin feature file
    Given a bugfix change in the spec phase
    When the scoped spec is written
    Then no .feature file is created in the change folder
    And spec.md contains no Gherkin Scenario blocks
