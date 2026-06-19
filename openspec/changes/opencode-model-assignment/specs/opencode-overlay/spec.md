# opencode-overlay Specification

## Purpose

Generate and merge an opencode agent overlay during `archon init` so the models in
`.archon/config.yaml` actually reach opencode. The overlay defines a primary
orchestrator plus one hidden subagent per SDD phase, with each subagent's `model`
injected per a decision tree, then is additively deep-merged into the shared
`~/.config/opencode/opencode.json`. All requirements below are Slice 1 (MVP) unless
tagged `[Slice 2]`.

## Requirements

### Requirement: Overlay Generation Gate

The overlay generation and merge step MUST run during `archon init` ONLY when the
resolved agent is `opencode`.

#### Scenario: opencode agent triggers overlay merge

- GIVEN `archon init` runs with the resolved agent equal to `opencode`
- WHEN init completes the skill extraction and template steps
- THEN the overlay is injected and merged into `~/.config/opencode/opencode.json`

#### Scenario: Non-opencode agent skips overlay

- GIVEN `archon init` runs with the resolved agent equal to `claude` or `codex`
- WHEN init runs
- THEN no overlay is generated AND `~/.config/opencode/opencode.json` is not touched

### Requirement: Orchestrator Agent Definition

The overlay MUST define an orchestrator agent under key `archon-orchestrator` with
`mode` set to `"primary"` and a prompt sourced from the generated `AGENTS.md`.

#### Scenario: Orchestrator entry present

- GIVEN the overlay is generated
- WHEN the `agent.archon-orchestrator` entry is inspected
- THEN `mode` is `"primary"`
- AND `prompt` references the generated `AGENTS.md` file

### Requirement: Delegation Allow-List

The `archon-orchestrator` agent MUST declare a `permission.task` allow-list that
permits delegation ONLY to the named SDD subagents and denies all others.

#### Scenario: Allow-list covers every SDD phase agent

- GIVEN the overlay is generated
- WHEN `agent.archon-orchestrator.permission.task` is inspected
- THEN it allows ALL of: `sdd-init`, `sdd-explore`, `sdd-propose`, `sdd-spec`,
  `sdd-design`, `sdd-tasks`, `sdd-apply`, `sdd-verify`, `sdd-archive`, `sdd-onboard`
- AND it denies the wildcard `*`

### Requirement: Per-Phase Subagent Definitions

The overlay MUST define one subagent per SDD phase. Each subagent MUST have `mode`
set to `"subagent"`, `hidden` set to `true`, a prompt pointing to its extracted skill
at `~/.config/opencode/skills/sdd-<phase>/SKILL.md`, and a `model` field resolved per
the injection decision tree.

#### Scenario: Subagent shape and skill prompt

- GIVEN the overlay is generated
- WHEN any `agent.sdd-<phase>` entry is inspected
- THEN `mode` is `"subagent"` AND `hidden` is `true`
- AND `prompt` points to `~/.config/opencode/skills/sdd-<phase>/SKILL.md`
- AND a `model` field is present

#### Scenario: Subagent keys match extracted skill folders

- GIVEN the SDD phases are extracted to `~/.config/opencode/skills/sdd-<phase>/`
- WHEN the overlay subagent keys are compared to the skill folder names
- THEN each subagent key `sdd-<phase>` matches an extracted `sdd-<phase>` folder

### Requirement: Model Injection Decision Tree

For each overlay agent the system MUST resolve the `model` field in this priority
order: (1) an explicit per-phase assignment from `.archon/config.yaml`
`models.phases` wins; (2) else if the agent key already exists in the user's
`opencode.json`, its existing definition MUST be preserved (not overwritten);
(3) else the resolved default model MUST be injected so no agent silently inherits
the orchestrator's model.

#### Scenario: Explicit per-phase assignment wins

- GIVEN `.archon/config.yaml` sets `models.phases.apply` to a model name
- WHEN the overlay is injected
- THEN `agent.sdd-apply.model` is set to the resolved value of that phase model

#### Scenario: Existing user agent is preserved

- GIVEN the user's `opencode.json` already contains `agent.sdd-spec`
- AND `.archon/config.yaml` has no explicit `models.phases.spec`
- WHEN the overlay is injected and merged
- THEN the user's existing `agent.sdd-spec` definition is preserved unchanged

#### Scenario: Default model fallback for new agents

- GIVEN a phase has no explicit assignment AND no pre-existing user agent entry
- WHEN the overlay is injected
- THEN `agent.sdd-<phase>.model` is set to the resolved default model
- AND it does NOT silently inherit the orchestrator's model

### Requirement: Additive Deep-Merge

The system MUST deep-merge the injected overlay into the SHARED
`~/.config/opencode/opencode.json`. The merge MUST be additive (recursive object
merge), preserving the user's pre-existing keys, except where a `__replace__`
sentinel forces whole-key replacement (used for the delegation `permission.task`
map).

#### Scenario: Pre-existing content preserved

- GIVEN `~/.config/opencode/opencode.json` already contains user agents and providers
- WHEN the overlay is deep-merged
- THEN all pre-existing user keys remain present
- AND the overlay's agents and settings are added alongside them

#### Scenario: Sentinel forces replacement

- GIVEN the overlay's `permission.task` map carries the `__replace__` sentinel
- WHEN the deep-merge runs against an existing `permission.task` map
- THEN the existing `permission.task` map is replaced wholesale, not merged

### Requirement: Backup and Rollback for Shared File

Because `opencode.json` is a shared global file, the system MUST record a pre-merge
backup in the rollback manifest. Rollback MUST restore the backup and MUST NEVER
delete the file. The file MUST NOT be added to `CreatedPaths` for deletion.

#### Scenario: Pre-merge backup recorded

- GIVEN `~/.config/opencode/opencode.json` exists before the merge
- WHEN `archon init` runs the merge
- THEN a backup of the pre-merge content is recorded in the rollback manifest

#### Scenario: Rollback restores backup without deleting

- GIVEN a backup of `opencode.json` is recorded in the rollback manifest
- WHEN `archon rollback` runs
- THEN the file is restored to its pre-merge content
- AND the file is NOT deleted

#### Scenario: No prior file before merge

- GIVEN no `~/.config/opencode/opencode.json` existed before the merge
- WHEN `archon rollback` runs
- THEN the file created by the merge may be removed
- AND no pre-existing user content is lost

### Requirement: Phase Models Documentation in AGENTS.md

The generated opencode `AGENTS.md` (`agentsTemplate`) MUST include a "Phase Models"
section documenting that per-phase models are wired via the `opencode.json` agent
definitions.

#### Scenario: AGENTS.md includes Phase Models section

- GIVEN `archon init` generates `AGENTS.md` for the opencode agent
- WHEN the rendered template is inspected
- THEN it contains a "Phase Models" section
- AND that section states per-phase models are wired via `opencode.json` agent
  definitions

### Requirement: Multi-Profile Overlays [Slice 2]

The system MUST support named SDD profiles, each generating its own per-phase
overlay/agents, with profile selection and re-application.

#### Scenario: Named profile generates its own agents

- GIVEN a named SDD profile is selected
- WHEN the overlay is generated
- THEN per-phase agents for that profile are produced and merged

### Requirement: Stale Agent Cleanup [Slice 2]

When re-applying overlays, the system MUST clean up stale archon-managed agents that
are no longer part of the active profile.

#### Scenario: Stale archon agent removed on re-apply

- GIVEN a previous overlay added an archon-managed agent no longer in the profile
- WHEN the overlay is re-applied
- THEN the stale archon-managed agent is removed
- AND user-authored agents are preserved

### Requirement: Re-apply via archon sync [Slice 2]

The system MUST provide an `archon sync` path that re-applies the overlay to
`opencode.json` after model edits via the TUI or `archon config set`.

#### Scenario: Edited models re-applied

- GIVEN a user edits `models.phases` via TUI or `archon config set`
- WHEN `archon sync` runs
- THEN the overlay is regenerated and merged with the updated resolved models
