Feature: Security coverage gate in verify and judge
  sdd-verify checks @security scenario coverage and harness-judge gates on it
  when the gate is on; no checks run when the gate is off.

  @happy @security
  Scenario: Full coverage passes verify
    Given security.enabled is true
    And every @security scenario has a covering test
    When sdd-verify runs
    Then the security coverage check passes

  @error @security
  Scenario: Uncovered abuse case fails verify
    Given security.enabled is true
    And one @security scenario has no covering test
    When sdd-verify runs
    Then the gap is reported as CRITICAL

  @error @security
  Scenario: Judge blocks on uncovered security scenario
    Given security.enabled is true
    And sdd-verify reported a CRITICAL security gap
    When harness-judge runs
    Then the judge gate fails

  @happy
  Scenario: No security gate when disabled
    Given security.enabled is false
    When sdd-verify and harness-judge run
    Then no @security coverage check is performed
