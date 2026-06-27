# Tasks Security Scanner Specification

## Purpose

When `security.enabled`, `sdd-tasks` MUST emit a tool-agnostic CI scanning task
covering SAST, secrets, and dependency vulnerabilities. When disabled, no scanning
task is emitted.

## Requirements

### Requirement: Tool-agnostic scanning task when enabled

When `security.enabled` is true, `sdd-tasks` MUST emit at least one task that runs a
security scan covering SAST, secret detection, and dependency vulnerability checks,
and that fails CI on any HIGH or CRITICAL finding. The task MUST NOT prescribe a
specific scanner tool. The task SHOULD be tagged `@security`.

#### Scenario: Scanning task is emitted when enabled

```gherkin
@happy @security
Scenario: Scanning task is emitted when enabled
  Given security.enabled is true
  When sdd-tasks generates the task list
  Then a task runs SAST, secret, and dependency scans
  And the task fails CI on any HIGH or CRITICAL finding
```

#### Scenario: Scanning task names no specific tool

```gherkin
@edge @security
Scenario: Scanning task names no specific tool
  Given security.enabled is true
  When the scanning task is generated
  Then the task references scan categories, not a named vendor tool
```

### Requirement: No scanning task when disabled

When `security.enabled` is false, `sdd-tasks` MUST NOT emit any security scanning task.
The task list MUST be identical to the pre-feature behavior.

#### Scenario: No scanning task when disabled

```gherkin
@happy
Scenario: No scanning task when disabled
  Given security.enabled is false
  When sdd-tasks generates the task list
  Then no security scanning task is present
```
