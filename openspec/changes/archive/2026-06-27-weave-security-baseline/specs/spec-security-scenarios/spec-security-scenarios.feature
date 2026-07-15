Feature: Security abuse-case scenarios in specs
  @security-tagged abuse cases authored in sdd-spec when the gate is on,
  and none emitted when the gate is off.

  @happy @security
  Scenario: Each MUST requirement gets an abuse case
    Given security.enabled is true
    And a spec with two MUST requirements
    When sdd-spec writes the feature files
    Then each MUST requirement has at least one @security-tagged scenario

  @edge @security
  Scenario: Abuse case describes the prohibited behavior
    Given security.enabled is true
    When an abuse-case scenario is authored
    Then it asserts the malicious action is rejected or contained

  @happy
  Scenario: No security tag when disabled
    Given security.enabled is false
    When sdd-spec writes the feature files
    Then no scenario carries the @security tag
