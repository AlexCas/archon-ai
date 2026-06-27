# Propose Security Risk Specification

## Purpose

When `security.enabled`, `sdd-propose` MUST add a mandatory security-risk row to the
proposal's Risks table, derived from the baseline module's risk taxonomy. When
disabled, proposal output is unchanged.

## Requirements

### Requirement: Mandatory security-risk row when enabled

When `security.enabled` is true, the `sdd-propose` output MUST include at least one
row in the Risks table that addresses a security risk, with likelihood and mitigation
filled in. The skill MUST load `skills/_shared/security-baseline.md` for the risk
taxonomy.

#### Scenario: Security risk row is emitted when enabled

```gherkin
@happy @security
Scenario: Security risk row is emitted when enabled
  Given security.enabled is true
  When sdd-propose produces a proposal
  Then the Risks table contains a security-risk row with likelihood and mitigation
```

#### Scenario: Missing security row is treated as incomplete

```gherkin
@error @security
Scenario: Missing security row is treated as incomplete
  Given security.enabled is true
  When a proposal omits any security-risk row
  Then the proposal is reported as incomplete
```

### Requirement: No proposal change when disabled

When `security.enabled` is false, `sdd-propose` MUST NOT add any security-risk row and
MUST NOT load the baseline module. Proposal output MUST be identical to the
pre-feature behavior.

#### Scenario: No security row when disabled

```gherkin
@happy
Scenario: No security row when disabled
  Given security.enabled is false
  When sdd-propose produces a proposal
  Then no security-specific row is added to the Risks table
```
