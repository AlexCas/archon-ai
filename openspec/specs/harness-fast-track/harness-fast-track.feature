Feature: Fast track
  A third track value, fast, offers a shorter phase sequence for small,
  well-understood feature work, selected only via explicit sdd-explore suggestion
  and user confirmation, never silently.

  @happy
  Scenario: Explicit fast track is honored
    Given a change whose state.yaml has track set to fast
    When the harness resolves the change track
    Then the fast 5-phase sequence applies
    And its spec phase is a full Gherkin spec, not the scoped bugfix format

  @happy
  Scenario: Fast advances explore to spec
    Given a fast change with explore completed
    When the orchestrator requests the spec phase
    Then the transition is allowed

  @error
  Scenario: Propose is not reachable in fast track
    Given a fast change with explore completed
    When the orchestrator requests the propose phase
    Then the transition is blocked
    And the response indicates propose is not part of the fast track

  @happy
  Scenario: Verify transitions straight to archive regardless of tier
    Given a fast change with verify completed
    And a computed risk tier that would require judge on the full track
    When the orchestrator requests the next phase
    Then the allowed next phase is archive
    And no judge phase is required or run

  @happy
  Scenario: Small classification suggests fast via one scoped question
    Given sdd-explore classifies the change size as small
    When the orchestrator prepares the track decision
    Then it surfaces exactly one scoped question suggesting track fast
    And it does not select fast without the user's confirmation

  @edge
  Scenario: Unknown classification defaults to full
    Given sdd-explore cannot classify the change size (unknown)
    When the orchestrator prepares the track decision
    Then the track defaults to full
    And fast is not suggested

  @happy
  Scenario: Human Review Gate fires once after spec
    Given a fast change producing a full spec
    When the spec phase completes
    Then the Human Review Gate fires for the spec
    And it does not fire again before apply
