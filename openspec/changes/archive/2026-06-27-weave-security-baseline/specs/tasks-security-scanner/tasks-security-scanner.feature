Feature: Security scanning task in tasks
  A tool-agnostic CI scanning task (SAST, secrets, dependency vulns) emitted
  by sdd-tasks when the gate is on, and none when the gate is off.

  @happy @security
  Scenario: Scanning task is emitted when enabled
    Given security.enabled is true
    When sdd-tasks generates the task list
    Then a task runs SAST, secret, and dependency scans
    And the task fails CI on any HIGH or CRITICAL finding

  @edge @security
  Scenario: Scanning task names no specific tool
    Given security.enabled is true
    When the scanning task is generated
    Then the task references scan categories, not a named vendor tool

  @happy
  Scenario: No scanning task when disabled
    Given security.enabled is false
    When sdd-tasks generates the task list
    Then no security scanning task is present
