Feature: Security baseline shared module
  A non-invocable, profile-scaled OWASP-derived checklist at
  skills/_shared/security-baseline.md, loaded by phase skills when enabled.

  @happy
  Scenario: CLI profile surfaces CLI-relevant controls
    Given security.profile is "cli"
    When a phase skill loads the baseline module
    Then the cli checklist includes injection, secret handling, and dependency integrity

  @happy
  Scenario: Web profile adds web Top 10 controls
    Given security.profile is "web"
    When a phase skill loads the baseline module
    Then the web checklist adds broken access control and SSRF controls

  @edge
  Scenario: Module is reference-only
    Given the security-baseline module
    Then it is referenced by path from phase skills
    And it is never invoked as a standalone skill
