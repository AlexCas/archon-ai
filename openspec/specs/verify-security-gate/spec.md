# Verify Security Gate Specification

## Purpose

When `security.enabled`, `sdd-verify` MUST check that every `@security` scenario has
covering verification, and `harness-judge` MUST gate on that coverage. When disabled,
no such checks run.

## Requirements

### Requirement: Verify checks @security coverage

When `security.enabled` is true, `sdd-verify` MUST confirm each `@security`-tagged
scenario maps to a covering test or scanner invocation. Any uncovered `@security`
scenario MUST be reported as CRITICAL.

#### Scenario: Full coverage passes verify

```gherkin
@happy @security
Scenario: Full coverage passes verify
  Given security.enabled is true
  And every @security scenario has a covering test
  When sdd-verify runs
  Then the security coverage check passes
```

#### Scenario: Uncovered abuse case fails verify

```gherkin
@error @security
Scenario: Uncovered abuse case fails verify
  Given security.enabled is true
  And one @security scenario has no covering test
  When sdd-verify runs
  Then the gap is reported as CRITICAL
```

### Requirement: Harness-judge security gate

When `security.enabled` is true, `harness-judge` MUST treat unresolved `@security`
coverage gaps as a failing gate, blocking archive until resolved. When
`security.enabled` is false, neither verify nor judge MUST run any security check and
behavior MUST be unchanged.

#### Scenario: Judge blocks on uncovered security scenario

```gherkin
@error @security
Scenario: Judge blocks on uncovered security scenario
  Given security.enabled is true
  And sdd-verify reported a CRITICAL security gap
  When harness-judge runs
  Then the judge gate fails
```

#### Scenario: No security gate when disabled

```gherkin
@happy
Scenario: No security gate when disabled
  Given security.enabled is false
  When sdd-verify and harness-judge run
  Then no @security coverage check is performed
```
