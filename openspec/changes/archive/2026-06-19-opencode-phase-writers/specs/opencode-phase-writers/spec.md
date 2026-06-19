# opencode-phase-writers Specification

## Purpose

For an opencode project, `archon init` and the TUI save path MUST write one `archon-<phase>`
subagent into `opencode.json` per resolvable SDD phase, each carrying its resolved per-phase
model as a `provider/model` FullID, alongside the existing `archon-leader`. Resolution reuses
`config.ResolvePhaseModels`, so the opencode subagents agree with the AGENTS.md "Phase Models"
advisory. The leader's shape and behavior are unchanged.

## Requirements

### Requirement: Per-phase subagent emission

The writer MUST emit one subagent keyed `archon-<phase>` for each phase returned by
`config.ResolvePhaseModels(cfg.Models)` (chain `Phases[phase]`→`Default`, omit-when-empty;
`judge` is never in PhaseOrder). A phase with no resolvable model MUST NOT be written.

#### Scenario: Init writes a subagent per resolvable phase

```gherkin
Scenario: Init writes a subagent per resolvable phase
  Given an opencode project with models.default set and models.phases.spec set
  When the user runs "archon init --agent opencode"
  Then "opencode.json" has an "agent.archon-<phase>" for every phase ResolvePhaseModels returns
  And no "agent.archon-judge" entry is written
```

#### Scenario: A phase with no resolvable model is omitted

```gherkin
@edge
Scenario: A phase with no resolvable model is omitted
  Given an opencode project with models.leader set but models.default and all models.phases empty
  When the user runs "archon init --agent opencode"
  Then no "agent.archon-<phase>" subagent is written
  And "agent.archon-leader" is written
```

### Requirement: Per-phase model is the resolved FullID

Each `archon-<phase>` subagent's `model` MUST equal the `PhaseModel.Model` (FullID) that
`ResolvePhaseModels` produces for that phase — identical to the value in the AGENTS.md advisory.

#### Scenario: Subagent model matches the resolved FullID

```gherkin
Scenario: Subagent model matches the resolved FullID
  Given models.phases.spec is provider "opencode" model "deepseek-v4-pro"
  When init writes the opencode agents
  Then "agent.archon-spec.model" equals "opencode/deepseek-v4-pro"
```

#### Scenario: Phase falls back to default model

```gherkin
@edge
Scenario: Phase falls back to default model
  Given models.phases.tasks is empty and models.default is provider "anthropic" model "claude-sonnet-4-6"
  When init writes the opencode agents
  Then "agent.archon-tasks.model" equals "anthropic/claude-sonnet-4-6"
```

### Requirement: Fixed per-phase subagent shape

Each `archon-<phase>` subagent MUST be written with these fields in this deterministic order:
`mode:"subagent"`, `hidden:true`, `model`, `description` ("Archon SDD <phase> phase"),
`prompt` ("{file:./AGENTS.md}").

#### Scenario: Subagent carries the fixed fields

```gherkin
Scenario: Subagent carries the fixed fields
  Given an opencode project with a resolvable phase model
  When init writes the opencode agents
  Then every "agent.archon-<phase>" has "mode" == "subagent" and "hidden" == true
  And its "description" == "Archon SDD <phase> phase" and "prompt" == "{file:./AGENTS.md}"
```

### Requirement: Whole-merge no-op condition

The merge MUST be a no-op (no `opencode.json` created or modified) when the leader FullID is
empty AND `ResolvePhaseModels` returns no phases, or when the agent is not opencode. Otherwise
the leader (when its FullID is non-empty) plus all resolvable phase subagents are written.

#### Scenario: Nothing configured writes nothing

```gherkin
@error
Scenario: Nothing configured writes nothing
  Given an opencode project with empty leader, empty default, and all phases empty
  When the user runs "archon init --agent opencode"
  Then no "opencode.json" is created or modified
```

#### Scenario: Phases set but empty leader still writes subagents

```gherkin
@edge
Scenario: Phases set but empty leader still writes subagents
  Given an opencode project with empty leader and models.default set
  When the user runs "archon init --agent opencode"
  Then "opencode.json" is written with the resolvable phase subagents
  And no "agent.archon-leader" entry is written
```

### Requirement: Deterministic output and preservation

The written `opencode.json` MUST be deterministic and idempotent (sorted map keys, fixed struct
field order, trailing newline, atomic temp+rename) and MUST preserve pre-existing top-level keys
and user-defined agents; only `archon-leader` and `archon-<phase>` entries are set.

#### Scenario: Re-run is byte-identical and preserves user agents

```gherkin
@edge
Scenario: Re-run is byte-identical and preserves user agents
  Given an opencode.json with an unrelated user agent and archon agents already written
  When init runs again with the same configuration
  Then the resulting opencode.json is byte-identical to the previous run
  And the pre-existing user agent and top-level keys are unchanged
```
