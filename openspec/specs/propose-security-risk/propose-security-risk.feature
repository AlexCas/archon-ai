Feature: Security risk in proposals
  A mandatory security-risk row in sdd-propose when the gate is on,
  and zero proposal change when the gate is off.

  @happy @security
  Scenario: Security risk row is emitted when enabled
    Given security.enabled is true
    When sdd-propose produces a proposal
    Then the Risks table contains a security-risk row with likelihood and mitigation

  @error @security
  Scenario: Missing security row is treated as incomplete
    Given security.enabled is true
    When a proposal omits any security-risk row
    Then the proposal is reported as incomplete

  @happy
  Scenario: No security row when disabled
    Given security.enabled is false
    When sdd-propose produces a proposal
    Then no security-specific row is added to the Risks table
