Feature: Session status resume file
  A SESSION_STATUS.md at the repo root tracks live session state so work can
  resume without losing context.

  Scenario: File updated on a phase transition
    Given an active change in the spec phase
    When the orchestrator advances to the design phase
    Then "SESSION_STATUS.md" at the repo root reflects phase "design"
    And it lists the completed phases with timestamps

  Scenario: Resuming after an unexpected close
    Given a "SESSION_STATUS.md" left at the repo root from a previous session
    When a new session starts
    Then the orchestrator reads it before evaluating any phase transition

  Scenario: Session status archived with the feature
    Given a completed change being archived in openspec mode
    When sdd-archive runs
    Then "SESSION_STATUS.md" is moved into the archived change folder
    And no "SESSION_STATUS.md" remains at the repo root
