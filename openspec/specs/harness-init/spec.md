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
