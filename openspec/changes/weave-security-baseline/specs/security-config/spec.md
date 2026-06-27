# Security Config Specification

## Purpose

Defines the opt-in `security` configuration surface: a Go config block, CLI get/set,
an `archon init --security` flag, and a TUI Security tab. Mirrors the existing
`playwright` config. Default is OFF so existing projects see zero behavior change.

## Requirements

### Requirement: Security config block

The system MUST expose a `security` config block with `enabled` (bool) and `profile`
(string) fields. `enabled` MUST default to `false` when the block is absent. The
config copy operation (`Clone()`) MUST preserve both fields.

#### Scenario: Absent block defaults to disabled

```gherkin
Scenario: Absent block defaults to disabled
  Given a config file with no security block
  When the config is loaded
  Then security.enabled is false
  And security.profile is empty
```

#### Scenario: Clone preserves security fields

```gherkin
Scenario: Clone preserves security fields
  Given a config with security.enabled true and profile "web"
  When the config is cloned
  Then the clone reports security.enabled true and profile "web"
```

### Requirement: Profile values restricted to cli and web

The system MUST accept only `cli` or `web` as `security.profile`. Any other value
MUST be rejected with an error. The `llm` and `agentic` profiles MUST NOT be accepted.

#### Scenario: Setting a valid profile succeeds

```gherkin
@happy
Scenario: Setting a valid profile succeeds
  When the user runs "archon config set security.profile web"
  Then the command succeeds
  And "archon config get security.profile" returns "web"
```

#### Scenario: Setting an invalid profile is rejected

```gherkin
@error @security
Scenario: Setting an invalid profile is rejected
  When the user runs "archon config set security.profile llm"
  Then the command fails with a non-zero exit code
  And the error names the supported values cli and web
```

### Requirement: Init flag and TUI tab enable security

The system MUST provide an `archon init --security` flag that sets `security.enabled`
to `true` in the emitted config. The TUI MUST expose a Security tab to toggle
`enabled` and select `profile`.

#### Scenario: Init flag enables the gate

```gherkin
@happy
Scenario: Init flag enables the gate
  When the user runs "archon init --security"
  Then the emitted .archon/config.yaml contains security.enabled true
```

#### Scenario: Init without flag leaves security off

```gherkin
@happy
Scenario: Init without flag leaves security off
  When the user runs "archon init" without --security
  Then the emitted config has security.enabled false
```
