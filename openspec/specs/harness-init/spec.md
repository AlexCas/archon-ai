# harness-init Specification

## Purpose

Bootstrap the harness in a project regardless of pre-existing agent folders, protect
hand-written orchestrator files, and configure web testing and models at init time.

## Requirements

### Requirement: Agent folder creation

`archon init` MUST create the selected agent's folder (`.claude`, `.opencode`,
`.agents`, `.codex`) when it does not already exist, instead of failing. An explicit
`--agent` selection MUST NOT require the folder to pre-exist. When no agent is given
and none is detected, init MUST report an error telling the user to pass `--agent`.

#### Scenario: Blank project initialized with an explicit agent

```gherkin
Scenario: Blank project initialized with an explicit agent
  Given a project directory with no agent folder
  When the user runs "archon init --agent claude"
  Then the ".claude" directory is created
  And ".archon/config.yaml" is created
  And "CLAUDE.md" is created
```

### Requirement: Existing orchestrator file guard

If the orchestrator file (`CLAUDE.md` for claude, `AGENTS.md` otherwise) already
exists, init MUST NOT overwrite it silently. It MUST prompt the user to replace it.
If the user declines, init MUST NOT perform ANY changes. `--force` MUST replace
without prompting.

#### Scenario: User declines to replace an existing orchestrator file

```gherkin
Scenario: User declines to replace an existing orchestrator file
  Given a project with an existing "CLAUDE.md"
  When the user runs "archon init --agent claude" and answers "n"
  Then the existing "CLAUDE.md" is left unchanged
  And no ".archon" directory is created
```

#### Scenario: Force replaces without prompting

```gherkin
Scenario: Force replaces without prompting
  Given a project with an existing "CLAUDE.md"
  When the user runs "archon init --agent claude --force"
  Then "CLAUDE.md" is regenerated
  And init completes successfully
```

### Requirement: Playwright init switch

Init MUST accept a `--playwright` flag that sets `playwright.enabled` in
`.archon/config.yaml`, mirroring the mutation-testing switch. The value MUST be
editable later via `archon config set playwright.enabled` and via the TUI.

#### Scenario: Enabling Playwright at init

```gherkin
Scenario: Enabling Playwright at init
  Given a web project
  When the user runs "archon init --agent claude --playwright"
  Then "playwright.enabled" is true in .archon/config.yaml
```

### Requirement: Dynamic model selection with free-form fallback

Model configuration MUST offer a model catalog that is DYNAMIC: it reflects the
agent CLIs detected on the system `PATH`. Models belonging to an agent CLI that is
NOT installed MUST be hidden from the offered/cyclable catalog. When the `opencode`
CLI is installed, the opencode portion of the catalog MUST be enumerated live from
that CLI instead of the static curated list. ONLY when `opencode` IS installed but
live enumeration fails (error, timeout, or unparseable output) MUST the offered
catalog fall back silently to the curated opencode list; detection MUST NEVER break
or block the TUI. (An absent `opencode` CLI does NOT trigger the curated fallback —
its models are hidden per the filter rule above.) Detection MUST
run once when the Models view is opened and be cached for that session — NOT per
keystroke and NOT during `archon init`. Claude and any remote-only provider stay
curated (no local enumeration). Free-form entry MUST always remain accepted (any
string), and `NormalizeModel`, `Validate`, and free-form acceptance behavior MUST be
unchanged by this feature.

| Aspect | Behavior |
|--------|----------|
| Catalog source | Detected agent CLIs on PATH (cached once per Models view) |
| Uninstalled agent | Its models hidden from the offered catalog |
| opencode installed | Live-enumerated catalog replaces curated opencode list |
| Enumeration failure | opencode installed but enumeration fails/times out/unparseable → silent fallback to curated opencode list; TUI never blocked |
| Claude / remote-only | Stays curated; no local enumeration |
| Free-form / Validate / NormalizeModel | Unchanged; any string accepted with advisory warning |

#### Scenario: Installed opencode shows the live catalog

```gherkin
@happy
Scenario: Installed opencode shows the live catalog
  Given the "opencode" CLI is installed on PATH
  When the user opens the Models view
  Then the offered opencode models are enumerated live from the opencode CLI
  And the stale curated opencode list is not shown
```

#### Scenario: Only installed agents' models are offered

```gherkin
@happy
Scenario: Only installed agents' models are offered
  Given the "opencode" CLI is not installed on PATH
  When the user cycles the catalog in the Models view
  Then no opencode models appear in the offered catalog
  And models for installed agents remain offered
```

#### Scenario: Detection is cached once per Models view

```gherkin
@edge
Scenario: Detection is cached once per Models view
  Given the Models view has been opened and detection has run once
  When the user cycles models and types repeatedly
  Then detection does not run again for that session
  And it never runs during "archon init"
```

#### Scenario: Live enumeration error falls back silently

```gherkin
@error
Scenario: Live enumeration error falls back silently
  Given the "opencode" CLI is installed but enumeration fails, times out, or returns unparseable output
  When the user opens the Models view
  Then the curated opencode list is offered as a silent fallback
  And the TUI is neither blocked nor shown an error
```

#### Scenario: Free-form entry and advisory behavior unchanged

```gherkin
@happy
Scenario: Free-form entry and advisory behavior unchanged
  Given the Models view default model input is empty
  When the user types "some-custom-model"
  Then the value is accepted and a non-blocking warning may be shown
  And NormalizeModel and Validate behave exactly as before this feature
```

### Requirement: Truthful skill inventory versions

The `skill_inventory` written to `.archon/config.yaml` at init MUST record each skill's
real version, read from that skill's `SKILL.md` `metadata.version` frontmatter. Init
MUST NOT write a hardcoded version. The shared skill-refresh routine that builds this
inventory is reused by `archon update`.

#### Scenario: Init records real frontmatter versions

```gherkin
Scenario: Init records real frontmatter versions
  Given embedded skills whose "SKILL.md" frontmatter declares versions like "2.0" and "3.0"
  When the user runs "archon init --agent claude"
  Then each "skill_inventory" entry records that skill's real frontmatter version
  And no inventory entry uses a hardcoded version
```

#### Scenario: Missing frontmatter version is handled

```gherkin
@edge
Scenario: Missing frontmatter version is handled
  Given an embedded skill whose "SKILL.md" declares no metadata.version
  When the user runs "archon init --agent claude"
  Then that skill is still recorded in "skill_inventory"
  And init does not abort
```

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
