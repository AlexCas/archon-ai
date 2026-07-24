# Delta for openspec-convention

## ADDED Requirements

### Requirement: Vault Root Shape Documentation

`skills/_shared/openspec-convention.md` MUST document the vault root shape —
`openspec/map.md` as the entry node, `specs/`, and `changes/` (with `archive/`) — as
the canonical reference all phase skills inherit. It MUST include a pointer to
`spec-vault` as the authoritative link-convention source.

#### Scenario: Convention documents map.md as entry node

```gherkin
@happy
Scenario: Convention documents map.md as entry node
  Given the openspec-convention shared module
  When a phase skill reads the convention for directory structure
  Then it finds openspec/map.md listed as the vault entry node
  And a pointer to the spec-vault convention for link rules
```

### Requirement: Hybrid Link Convention in Convention Doc

`openspec-convention.md` MUST document the hybrid link convention: wikilinks for
capability-identity references and relative links for intra-change navigation. It MUST
specify which link form applies to which reference category so phase skills have an
unambiguous rule to follow.

#### Scenario: Phase skill reads unambiguous link rule

```gherkin
@happy
Scenario: Phase skill reads unambiguous link rule
  Given the openspec-convention shared module
  When a phase skill decides how to link to a capability
  Then the convention specifies [[capability]] wikilink for capability-identity references
  And relative markdown links for intra-change artifact navigation
```

### Requirement: Artifact File Paths Table Updated

The artifact file paths table in `openspec-convention.md` MUST be extended to include
`sdd-init` creating `openspec/map.md` and `archon map` regenerating it on phase
transitions and archive.

#### Scenario: Table lists map.md creation and regen

```gherkin
@happy
Scenario: Table lists map.md creation and regen
  Given the artifact file paths table in openspec-convention.md
  When a skill author reads the table
  Then it shows sdd-init creates openspec/map.md
  And archon map regenerates openspec/map.md after every phase transition and archive
```
