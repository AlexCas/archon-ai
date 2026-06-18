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

### Requirement: Static model selection with free-form fallback

Model configuration MUST offer a curated static catalog of Claude models and Opencode
Go models as selectable options, AND MUST still accept any free-form model string.
Unknown models are accepted with a warning, never rejected.

#### Scenario: Selecting a static model in the TUI

```gherkin
Scenario: Selecting a static model in the TUI
  Given the Models tab is focused on the default model input
  When the user cycles the static catalog with ctrl+n
  Then the input is set to a Claude or Opencode Go model from the catalog

Scenario: Typing a free-form model
  Given the Models tab default model input is empty
  When the user types "some-custom-model"
  Then the value is accepted and a non-blocking warning may be shown
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
before rendering. Normalization MUST recognize curated identifiers from FOUR providers
— Claude, Gemini, OpenAI, Opencode — with canonical output per provider:

| Provider | Match source | Canonical output |
|----------|--------------|------------------|
| Claude | `opus` / `sonnet` / `haiku` family token | short family alias (`opus`/`sonnet`/`haiku`) |
| Gemini | curated `GeminiModels` catalog id | matched catalog id, as-is |
| OpenAI | curated `OpenAIModels` catalog id | matched catalog id, as-is |
| Opencode | curated `OpencodeModels` catalog id | matched catalog id, as-is |

Matching MUST be whole-token: a value merely CONTAINING a family substring (e.g.
`octopus`) MUST NOT match `opus`. Raw display strings (e.g. "Opus 4.8") MUST NOT appear
in the block. When a value normalizes for ANY provider, the phase line renders;
otherwise it returns not-ok, its phase is OMITTED, and the value is accepted with an
advisory warning, never rejected. Catalog contents are a design detail.

#### Scenario: Display name is normalized to an accepted identifier

```gherkin
Scenario: Display name is normalized to an accepted identifier
  Given "models.phases.design" is set to a display string like "Opus 4.8"
  When the orchestrator template is rendered
  Then the "design" line shows a normalized identifier the delegation tool accepts
  And no raw display string appears in the block
```

#### Scenario: Gemini model normalizes to its catalog id

```gherkin
Scenario: Gemini model normalizes to its catalog id
  Given "models.phases.spec" is set to a curated Gemini catalog id
  When the orchestrator template is rendered
  Then the "spec" line shows that Gemini catalog id as-is
```

#### Scenario: OpenAI model normalizes to its catalog id

```gherkin
Scenario: OpenAI model normalizes to its catalog id
  Given "models.phases.tasks" is set to a curated OpenAI catalog id
  When the orchestrator template is rendered
  Then the "tasks" line shows that OpenAI catalog id as-is
```

#### Scenario: Opencode model normalizes to its catalog id

```gherkin
Scenario: Opencode model normalizes to its catalog id
  Given "models.phases.apply" is set to a curated Opencode catalog id
  When the orchestrator template is rendered
  Then the "apply" line shows that Opencode catalog id as-is
```

#### Scenario: Whole-token guard rejects a containing substring

```gherkin
@edge
Scenario: Whole-token guard rejects a containing substring
  Given "models.phases.verify" is set to "octopus"
  When the value is normalized
  Then it does not match the Claude "opus" family
```

#### Scenario: Unresolvable typo is omitted but not rejected

```gherkin
@error
Scenario: Unresolvable typo is omitted but not rejected
  Given "models.phases.propose" is set to an unresolvable value like "Opues 4.8"
  When the configured models are processed for rendering
  Then "propose" is omitted from the block
  And the value is accepted with an advisory warning, not rejected
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

### Requirement: Cross-provider normalization precedence

When a value could match more than one provider, normalization MUST resolve it by a
fixed precedence: Claude → Gemini → OpenAI → Opencode. The first provider whose
whole-token match succeeds wins. This precedence MUST be stable across runs so the
`## Phase Models` block is byte-identical between the `archon init` and TUI regenerate
paths.

#### Scenario: Colliding value resolves by fixed precedence

```gherkin
@edge
Scenario: Colliding value resolves by fixed precedence
  Given a value that matches both Claude and a later provider
  When the value is normalized
  Then it resolves to the Claude canonical form
```

#### Scenario: Non-Claude default renders an identical block across paths

```gherkin
@happy
Scenario: Non-Claude default renders an identical block across paths
  Given "models.default" is set to a curated non-Claude catalog id
  When the file is rendered via "archon init" and via the TUI regenerate path
  Then both produce a non-empty "## Phase Models" block
  And the two blocks are byte-identical
```
