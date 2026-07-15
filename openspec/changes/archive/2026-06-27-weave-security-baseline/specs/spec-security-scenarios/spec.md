# Spec Security Scenarios Specification

## Purpose

When `security.enabled`, `sdd-spec` MUST author `@security`-tagged abuse-case Gherkin
scenarios alongside the normal scenarios for each MUST requirement. When disabled, no
`@security` scenarios are emitted.

## Requirements

### Requirement: Abuse-case scenarios when enabled

When `security.enabled` is true, `sdd-spec` MUST derive at least one `@security`-tagged
abuse-case scenario for each MUST requirement in the change. Prohibitions in those
scenarios MUST use RFC 2119 `MUST NOT` phrasing in the parent requirement.

#### Scenario: Each MUST requirement gets an abuse case

```gherkin
@happy @security
Scenario: Each MUST requirement gets an abuse case
  Given security.enabled is true
  And a spec with two MUST requirements
  When sdd-spec writes the feature files
  Then each MUST requirement has at least one @security-tagged scenario
```

#### Scenario: Abuse case describes the prohibited behavior

```gherkin
@edge @security
Scenario: Abuse case describes the prohibited behavior
  Given security.enabled is true
  When an abuse-case scenario is authored
  Then it asserts the malicious action is rejected or contained
```

### Requirement: No security scenarios when disabled

When `security.enabled` is false, `sdd-spec` MUST NOT emit any `@security`-tagged
scenario. Generated feature files MUST be identical to the pre-feature behavior.

#### Scenario: No security tag when disabled

```gherkin
@happy
Scenario: No security tag when disabled
  Given security.enabled is false
  When sdd-spec writes the feature files
  Then no scenario carries the @security tag
```
