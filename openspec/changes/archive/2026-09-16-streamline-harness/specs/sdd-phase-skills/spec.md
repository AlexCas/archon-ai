# Delta for sdd-phase-skills

<!-- [[sdd-phase-skills]] · [proposal](../../proposal.md) · [exploration](../../exploration.md) -->

This delta makes the phase skills OpenSpec-only (removing engram/hybrid/none persistence
branches) and removes the Impeccable and Graphify conditional annotation hooks from the
phase skills. The existing wikilink/relative-link requirements are unchanged.

## ADDED Requirements

### Requirement: Phase skills persist to OpenSpec only

The phase skills (`sdd-explore`, `sdd-propose`, `sdd-spec`, `sdd-design`, `sdd-tasks`,
`sdd-apply`, `sdd-verify`, `sdd-archive`) MUST persist and read artifacts through OpenSpec
files only. They MUST NOT branch on an artifact-store mode and MUST NOT contain
`engram`, `hybrid`, or `none` persistence paths. The shared persistence contract
(`skills/_shared/persistence-contract.md`) MUST describe OpenSpec as the sole mode, and
the `skills/_shared/engram-convention.md` module MUST NOT exist.

#### Scenario: A phase skill has a single OpenSpec persistence path

```gherkin
@happy
Scenario: A phase skill has a single OpenSpec persistence path
  Given any SDD phase skill after this change is applied
  When its persistence contract is read
  Then it describes writing and reading OpenSpec files only
  And it contains no engram, hybrid, or none branch
```

#### Scenario: The engram convention module is gone

```gherkin
@happy
Scenario: The engram convention module is gone
  Given the harness after this change is applied
  When the shared skill modules are enumerated
  Then skills/_shared/engram-convention.md is not present
  And skills/_shared/persistence-contract.md names OpenSpec as the sole mode
```

#### Scenario: No none-mode "return only" path remains

```gherkin
@edge
Scenario: No none-mode "return only" path remains
  Given any SDD phase skill after this change is applied
  When its instructions are read
  Then there is no "none" mode that returns results without writing OpenSpec files
  And artifacts are always written to the change folder
```

### Requirement: No Impeccable or Graphify hooks in phase skills

The phase skills MUST NOT contain Impeccable or Graphify conditional blocks. `sdd-spec`
MUST NOT emit an `@design`/Impeccable annotation note. `sdd-explore` MUST NOT emit an
Impeccable recommendation or consume a Graphify code graph. `sdd-design`, `sdd-apply`,
`sdd-verify`, and `sdd-tasks` MUST NOT contain Impeccable/Graphify conditional steps.
Removing these hooks MUST NOT change any other phase-skill behavior (the Security
`@security` abuse-case behavior, gated by `security.enabled`, is unaffected).

#### Scenario: sdd-spec emits no Impeccable design note

```gherkin
@happy
Scenario: sdd-spec emits no Impeccable design note
  Given sdd-spec after this change is applied
  When it writes a spec for a frontend design-language requirement
  Then it does not add an @design or Impeccable annotation
  And no impeccable.enabled condition is evaluated
```

#### Scenario: sdd-explore has no Graphify consumption

```gherkin
@happy
Scenario: sdd-explore has no Graphify consumption
  Given sdd-explore after this change is applied
  When it maps the current state of a repository
  Then it does not shell any graphify command
  And it does not read a code-graph excerpt
```

#### Scenario: Security abuse-case behavior is unchanged

```gherkin
@edge
Scenario: Security abuse-case behavior is unchanged
  Given security.enabled is true after this change is applied
  When sdd-spec writes a spec for a MUST requirement
  Then it still derives at least one @security abuse-case scenario
  And the Impeccable/Graphify removal does not affect this behavior
```
