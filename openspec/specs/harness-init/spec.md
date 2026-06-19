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

### Requirement: Configurable leader model

`.archon/config.yaml` MUST hold a `models.leader` field carrying the FULL
`provider/model-id` value verbatim (e.g. `anthropic/claude-sonnet-4-...`), with no
prefix stripping or normalization applied to the stored value. The field MUST survive
`Config.Clone()` and config serialization round-trips. Any validation of the value is
advisory only and MUST NOT reject or rewrite it.

#### Scenario: Leader model survives clone and round-trip

```gherkin
Scenario: Leader model survives clone and round-trip
  Given "models.leader" set to "anthropic/claude-sonnet-4-20250514"
  When the config is cloned and serialized then reloaded
  Then "models.leader" equals "anthropic/claude-sonnet-4-20250514" verbatim
```

### Requirement: Opencode archon-leader agent merge at init

For the **opencode** agent, `archon init` MUST additively merge a primary agent named
`archon-leader` into the project `opencode.json` (creating the file when absent),
setting ONLY `agent.archon-leader` with: `mode: "primary"`, `prompt:
"{file:./AGENTS.md}"`, and `model` set to `models.leader`. The merge MUST be additive
and idempotent — it MUST NOT modify or remove unrelated keys, and re-running init MUST
produce the same result with no duplication or drift. The agent MUST NOT be set as the
default agent (no `default_agent` written). The written `opencode.json` path MUST be
registered in the rollback manifest.

#### Scenario: Init writes the archon-leader agent

```gherkin
Scenario: Init writes the archon-leader agent
  Given an opencode project with "models.leader" set and no "opencode.json"
  When the user runs "archon init --agent opencode"
  Then "opencode.json" sets "agent.archon-leader" with mode "primary"
  And its "prompt" is "{file:./AGENTS.md}" and "model" equals "models.leader"
  And the "opencode.json" path is registered for rollback
```

#### Scenario: Merge into an existing opencode.json preserves other keys

```gherkin
@edge
Scenario: Merge into an existing opencode.json preserves other keys
  Given an opencode project whose "opencode.json" has unrelated keys and agents
  When the user runs "archon init --agent opencode"
  Then "agent.archon-leader" is added
  And every pre-existing key and agent is left unchanged
  And no "default_agent" key is written
```

#### Scenario: Re-running init is idempotent

```gherkin
@edge
Scenario: Re-running init is idempotent
  Given an opencode project already initialized with "archon-leader"
  When the user runs "archon init --agent opencode" again
  Then "agent.archon-leader" appears exactly once with the same content
  And no other key drifts
```

### Requirement: Leader merge no-op guards

The archon-leader merge MUST be a no-op when the agent is not opencode OR when
`models.leader` is empty: no `opencode.json` is created and nothing is written.

#### Scenario: Non-opencode agent writes no opencode.json

```gherkin
@edge
Scenario: Non-opencode agent writes no opencode.json
  Given a project initialized with "archon init --agent claude"
  When init completes
  Then no "opencode.json" is created or modified
```

#### Scenario: Empty leader model writes nothing

```gherkin
@error
Scenario: Empty leader model writes nothing
  Given an opencode project with an empty "models.leader"
  When the user runs "archon init --agent opencode"
  Then no "opencode.json" is created
  And no archon-leader agent is written
```

### Requirement: TUI and update write-path parity

For an opencode project, saving the leader-model field in the TUI Models tab MUST
produce the SAME archon-leader merge result as `archon init`, with no divergence
between the two write paths. `archon update` MUST NOT write or rewrite the opencode
agent (update stays skill-only).

#### Scenario: TUI save matches init merge result

```gherkin
@happy
Scenario: TUI save matches init merge result
  Given an opencode project and a chosen leader model
  When the leader-model field is saved via the TUI Models tab
  Then the resulting "agent.archon-leader" equals what "archon init" would produce
```

#### Scenario: Update leaves the opencode agent untouched

```gherkin
@edge
Scenario: Update leaves the opencode agent untouched
  Given an opencode project with an existing "agent.archon-leader"
  When the user runs "archon update"
  Then "opencode.json" is not written or rewritten
```
