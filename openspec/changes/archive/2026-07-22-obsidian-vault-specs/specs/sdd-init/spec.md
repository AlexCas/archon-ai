# Delta for sdd-init

## ADDED Requirements

### Requirement: Seed map.md on Init

`sdd-init` / `archon init` MUST create `openspec/map.md` when scaffolding a new
`openspec/` directory. The seeded file MUST contain the managed markers
`<!-- MAP:START -->` … `<!-- MAP:END -->` with an empty or minimal generated region,
and MAY contain a brief authored preamble outside the markers. If `openspec/map.md`
already exists, init MUST NOT overwrite it.

#### Scenario: Init creates map.md with managed markers

```gherkin
@happy
Scenario: Init creates map.md with managed markers
  Given a project with no existing openspec/ directory
  When the user runs archon init
  Then openspec/map.md is created
  And it contains <!-- MAP:START --> and <!-- MAP:END --> markers
```

#### Scenario: Init does not overwrite an existing map.md

```gherkin
@edge
Scenario: Init does not overwrite an existing map.md
  Given a project where openspec/map.md already exists with authored content
  When the user runs archon init again
  Then openspec/map.md is left unchanged
```
