# Delta for sdd-phase-skills (sdd-propose, sdd-spec, sdd-design, sdd-tasks, sdd-verify)

## ADDED Requirements

### Requirement: Capability Wikilink Emission

The phase skills `sdd-propose`, `sdd-spec`, `sdd-design`, `sdd-tasks`, and `sdd-verify`
MUST emit `[[capability]]` wikilinks in their artifact bodies whenever they reference a
capability by name. Each wikilink MUST use the canonical capability identifier (the name
of the `specs/{capability}/` directory). Phase skills MUST NOT use bare prose names for
capability references in the artifact body.

#### Scenario: Proposal artifact contains wikilinks for referenced capabilities

```gherkin
@happy
Scenario: Proposal artifact contains wikilinks for referenced capabilities
  Given a proposal that references the harness-workflow and spec-vault capabilities
  When sdd-propose writes proposal.md
  Then proposal.md contains [[harness-workflow]] and [[spec-vault]] as wikilinks
  And not bare prose like "harness-workflow" without link syntax
```

#### Scenario: Spec artifact links to proposal with a relative link

```gherkin
@happy
Scenario: Spec artifact links to proposal with a relative link
  Given a spec artifact being written inside changes/my-feature/specs/foo/
  When sdd-spec writes spec.md
  Then spec.md contains a relative link to [proposal](../../proposal.md)
```

#### Scenario: Design artifact links to referenced capability specs with wikilinks

```gherkin
@happy
Scenario: Design artifact links to referenced capability specs with wikilinks
  Given a design artifact that references the archon-map capability
  When sdd-design writes design.md
  Then design.md contains [[archon-map]] as a wikilink
```

### Requirement: Intra-Change Relative Navigation

Phase skills MUST use relative links for intra-change navigation between artifacts
within the same change folder (proposal ↔ spec ↔ design ↔ tasks ↔ verify-report).
These relative links MUST resolve correctly given the nesting depth of each artifact
within the change folder.

#### Scenario: Tasks artifact links to design with a relative link

```gherkin
@happy
Scenario: Tasks artifact links to design with a relative link
  Given a tasks artifact being written at changes/my-feature/tasks.md
  When sdd-tasks writes tasks.md
  Then tasks.md contains a relative link to [design](design.md)
```

#### Scenario: Verify report links to tasks with a relative link

```gherkin
@happy
Scenario: Verify report links to tasks with a relative link
  Given a verify-report artifact being written at changes/my-feature/verify-report.md
  When sdd-verify writes verify-report.md
  Then verify-report.md contains a relative link to [tasks](tasks.md)
```
