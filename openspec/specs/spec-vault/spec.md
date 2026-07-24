# spec-vault Specification

## Purpose

The `spec-vault` capability defines the Obsidian-style vault layout for `openspec/`, the
hybrid link convention (wikilinks for capability identity, relative links for intra-change
navigation), the managed-marker policy, and the rule that `.feature` files stay put and are
only referenced. It is the single source of truth all phase skills inherit.

## Requirements

### Requirement: Vault Root Shape

The `openspec/` directory MUST follow a fixed root shape: `map.md` at the root as the
entry node, `specs/{capability}/` as the capability source-of-truth tree, and
`changes/{change}/` (plus `changes/archive/`) for active and completed changes.
No other top-level files MAY be introduced without updating this spec.

#### Scenario: Vault root contains map.md and known subdirectories

```gherkin
@happy
Scenario: Vault root contains map.md and known subdirectories
  Given an initialized archon project
  When the openspec/ directory is inspected
  Then openspec/map.md exists as the entry node
  And openspec/specs/ exists as the capability source-of-truth tree
  And openspec/changes/ exists with an archive/ subdirectory
```

### Requirement: Hybrid Link Convention

Phase skills MUST use `[[capability]]` wikilinks for capability-identity references
(change → capabilities it touches; `map.md` → capabilities). Phase skills MUST use
relative links `[text](relative/path.md)` for intra-change navigation (proposal ↔
spec ↔ design ↔ tasks within the same change folder). No other link forms are
permitted for these reference categories.

#### Scenario: Capability reference uses a wikilink

```gherkin
@happy
Scenario: Capability reference uses a wikilink
  Given a sdd-propose artifact referencing the harness-workflow capability
  When the artifact is written
  Then the reference appears as [[harness-workflow]] in the markdown body
  And not as a relative path link
```

#### Scenario: Intra-change navigation uses relative links

```gherkin
@happy
Scenario: Intra-change navigation uses relative links
  Given a sdd-spec artifact inside changes/my-feature/specs/foo/
  When the spec links to the change's proposal
  Then the link is a relative markdown link like [proposal](../../proposal.md)
  And not a wikilink
```

#### Scenario: Wikilink survives the archive move without rewrite

```gherkin
@happy
Scenario: Wikilink survives the archive move without rewrite
  Given a change artifact containing [[harness-workflow]]
  When sdd-archive moves the change to changes/archive/YYYY-MM-DD-change/
  Then the wikilink still reads [[harness-workflow]] with no modification
```

### Requirement: Managed-Marker Policy

Tooling (the `archon map` Go step) MUST write ONLY inside
`<!-- MAP:START -->` … `<!-- MAP:END -->` regions. Authored prose outside these markers
MUST NOT be modified by any automated tool. A document MAY contain at most one managed
region pair. Nested managed regions are PROHIBITED.

#### Scenario: Tooling writes only inside managed markers

```gherkin
@happy
Scenario: Tooling writes only inside managed markers
  Given openspec/map.md with authored prose outside managed markers
  When archon map regenerates map.md
  Then the content between <!-- MAP:START --> and <!-- MAP:END --> is replaced
  And all prose outside the markers is byte-identical to before
```

#### Scenario: Authored prose outside markers is never touched

```gherkin
@edge
Scenario: Authored prose outside markers is never touched
  Given openspec/map.md with custom authored sections outside the managed region
  When archon map is run multiple times
  Then no character outside the managed region is modified
```

### Requirement: Feature Files Stay Put

The `.feature` files beside `spec.md` MUST NOT be moved, renamed, or have their
content changed by any vault or link-convention operation. Markdown artifacts MAY
reference `.feature` files with relative links, but `.feature` file content and
location are managed exclusively by the SDD spec phase.

#### Scenario: Archive move does not touch feature files

```gherkin
@happy
Scenario: Archive move does not touch feature files
  Given a change with specs/foo/foo.feature
  When sdd-archive moves the change folder
  Then foo.feature is moved alongside spec.md to the archive location
  And foo.feature content is byte-identical to before the move
```
