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
