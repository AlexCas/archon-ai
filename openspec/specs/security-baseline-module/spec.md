# Security Baseline Module Specification

## Purpose

Defines `skills/_shared/security-baseline.md`, a non-invocable shared reference
module holding a profile-scaled, OWASP-derived secure-by-design checklist. Phase
skills load it conditionally when `security.enabled`.

## Requirements

### Requirement: Profile-scaled OWASP checklist

The module MUST provide an OWASP-derived control checklist scoped to two profiles,
`cli` and `web`. The `cli` profile MUST cover at least argument/path injection,
secret handling, and dependency integrity. The `web` profile MUST additionally cover
the OWASP Top 10 web categories (e.g. injection, broken access control, SSRF). It
MUST NOT contain `llm` or `agentic` sections.

#### Scenario: CLI profile surfaces CLI-relevant controls

```gherkin
@happy
Scenario: CLI profile surfaces CLI-relevant controls
  Given security.profile is "cli"
  When a phase skill loads the baseline module
  Then the cli checklist includes injection, secret handling, and dependency integrity
```

#### Scenario: Web profile adds web Top 10 controls

```gherkin
@happy
Scenario: Web profile adds web Top 10 controls
  Given security.profile is "web"
  When a phase skill loads the baseline module
  Then the web checklist adds broken access control and SSRF controls
```

### Requirement: Non-invocable shared module embedded automatically

The module MUST be a pure reference package — not user-invocable and not
model-invocable directly. It MUST live at `skills/_shared/security-baseline.md` so the
existing `all:_shared` embed directive ships it without changes to `embed.go`.

#### Scenario: Module is reference-only

```gherkin
@edge
Scenario: Module is reference-only
  Given the security-baseline module
  Then it is referenced by path from phase skills
  And it is never invoked as a standalone skill
```
