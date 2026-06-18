# Delta for harness-init

## ADDED Requirements

### Requirement: Rendered phase→model block

The generated orchestrator file (`CLAUDE.md` for claude, `AGENTS.md` otherwise) MUST
contain a phase→model block derived from `config.Models`. The block is ADVISORY: it
instructs the orchestrator LLM to request the configured model when delegating each
SDD phase. Both the `archon init` render path and the TUI "regenerate template" path
MUST emit the same block from the same configured models.

#### Scenario: Init renders a phase model for a configured phase

```gherkin
Scenario: Init renders a phase model for a configured phase
  Given a config with "models.phases.propose" set
  When the user runs "archon init --agent claude"
  Then the generated "CLAUDE.md" contains a phase→model block
  And the block lists "propose" with its resolved model
```

#### Scenario: TUI regeneration produces the same block as init

```gherkin
Scenario: TUI regeneration produces the same block as init
  Given a config rendered once by "archon init"
  When the same config is regenerated via the TUI template path
  Then the regenerated orchestrator file contains an identical phase→model block
```

### Requirement: Normalization to real model IDs

Configured model values MUST be normalized to identifiers the delegation tool accepts
before being rendered into the block. Accepted target forms are aliases (`opus`,
`sonnet`, `haiku`) and full IDs (e.g. `claude-opus-4-8`, `claude-sonnet-4-6`,
`claude-haiku-4-5-20251001`). Raw display strings (e.g. "Opus 4.8") MUST NOT appear
in the rendered block. The exact mapping table is a design detail.

#### Scenario: Display name is normalized to an accepted identifier

```gherkin
Scenario: Display name is normalized to an accepted identifier
  Given "models.phases.design" is set to a display string like "Opus 4.8"
  When the orchestrator template is rendered
  Then the "design" line shows a normalized identifier the delegation tool accepts
  And no raw display string appears in the block
```

### Requirement: Phase model resolution and fallback

For each SDD phase, the rendered model MUST be resolved as: explicit
`models.phases.<phase>` if set; else `models.default` if set; else the phase line MUST
be OMITTED from the block. Resolution and rendering MUST NOT mutate the user's stored
config.

#### Scenario: Phase falls back to the default model

```gherkin
Scenario: Phase falls back to the default model
  Given "models.phases.verify" is unset
  And "models.default" is set
  When the template is rendered
  Then the "verify" line shows the normalized default model
```

#### Scenario: Phase omitted when no model resolves

```gherkin
@edge
Scenario: Phase omitted when no model resolves
  Given "models.phases.apply" is unset
  And "models.default" is unset
  When the template is rendered
  Then the block contains no line for "apply"
```

### Requirement: Deterministic phase ordering

Rendered phase lines MUST follow the canonical SDD phase order (explore → propose →
spec → design → tasks → apply → verify → archive), NOT Go map iteration order, so the
block renders identically across runs and golden tests stay stable.

#### Scenario: Multiple configured phases render in canonical order

```gherkin
Scenario: Multiple configured phases render in canonical order
  Given "models.phases" sets "archive", "explore", and "design"
  When the template is rendered twice
  Then both renders list the phases in canonical SDD order
  And the two renders are byte-identical
```

### Requirement: Unknown model values surfaced to the user

A configured model value that does not resolve to a known model MUST be surfaced to the
end user with actionable feedback. Whether the feedback is a hard rejection or an
advisory warning is a design decision and is not constrained here.

#### Scenario: Garbage model value is surfaced

```gherkin
@error
Scenario: Garbage model value is surfaced
  Given "models.phases.propose" is set to an unresolvable value like "Opues 4.8"
  When the configured models are processed for rendering
  Then the user receives actionable feedback identifying the unknown value
```
