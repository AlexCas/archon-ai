# opencode-phase-subagents Specification

## Purpose

For an **opencode** project, `archon init` and the TUI save path MUST write one
subagent per SDD phase into the project `opencode.json`, each carrying its resolved
per-phase model written verbatim, so opencode honors the per-phase models configured
in `.archon/config.yaml`. This capability extends the existing single `archon-leader`
writer; it does not change the leader's shape or behavior, model normalization, or any
model-resolution package.

## Requirements

### Requirement: Per-phase subagent emission

For an opencode project, init and TUI save MUST write one subagent keyed
`archon-<phase>` for each of the 8 phases in `config.PhaseOrder` (explore, propose,
spec, design, tasks, apply, verify, archive). The `judge` phase MUST NOT be emitted.
Whenever the merge runs, all 8 `archon-<phase>` subagents MUST be present.

#### Scenario: Init writes 8 per-phase subagents

```gherkin
@happy
Scenario: Init writes 8 per-phase subagents
  Given an opencode project with per-phase models configured for every phase
  When the user runs "archon init --agent opencode"
  Then "opencode.json" contains an "agent.archon-<phase>" for each of the 8 PhaseOrder phases
  And each subagent's "model" equals the configured per-phase model verbatim
  And no "agent.archon-judge" entry is written
```

### Requirement: Verbatim per-phase model resolution

Each `archon-<phase>` subagent's `model` MUST be written verbatim, with no
normalization or prefix stripping. The model for a phase MUST be resolved by this
chain: `Models.Phases[phase]` when non-empty, else `Models.Default` when non-empty,
else `Models.Leader`.

#### Scenario: Phase with its own model uses it verbatim

```gherkin
@happy
Scenario: Phase with its own model uses it verbatim
  Given "models.phases.spec" set to "anthropic/claude-opus-4-20250514"
  When init writes the opencode agents
  Then "agent.archon-spec.model" equals "anthropic/claude-opus-4-20250514" verbatim
```

#### Scenario: Phase without a model falls back to default

```gherkin
@edge
Scenario: Phase without a model falls back to default
  Given "models.phases.tasks" is empty and "models.default" is "anthropic/claude-sonnet-4-20250514"
  When init writes the opencode agents
  Then "agent.archon-tasks" is still present
  And its "model" equals "models.default" verbatim
```

#### Scenario: Phase and default empty fall back to leader

```gherkin
@edge
Scenario: Phase and default empty fall back to leader
  Given "models.phases.verify" and "models.default" are both empty
  And "models.leader" is "anthropic/claude-opus-4-20250514"
  When init writes the opencode agents
  Then "agent.archon-verify" is still present
  And its "model" equals "models.leader" verbatim
```

### Requirement: Fixed per-phase subagent shape

Each `archon-<phase>` subagent MUST be written with these fields in this deterministic
order: `mode: "subagent"`, `hidden: true`, `model`, `description`, `prompt`. The
`description` and `prompt` MUST be static and identical across phases for deterministic
output; the `prompt` SHOULD be `"{file:./AGENTS.md}"`, mirroring the leader.

#### Scenario: Subagent carries the fixed fields

```gherkin
@happy
Scenario: Subagent carries the fixed fields
  Given an opencode project with any non-empty model configuration
  When init writes the opencode agents
  Then every "agent.archon-<phase>" has "mode" equal to "subagent"
  And every "agent.archon-<phase>" has "hidden" equal to true
  And every "agent.archon-<phase>" has a "model", a "description", and a "prompt"
```

### Requirement: Whole-merge no-op condition

The whole opencode merge MUST be a no-op (no `opencode.json` created or modified) only
when `Models.Leader == ""` AND `Models.Default == ""` AND every `Models.Phases` value
is empty. The merge MUST also be a no-op when the agent is not opencode. Otherwise the
merge MUST write the leader (when leader is non-empty, with its existing unchanged
shape) plus all 8 `archon-<phase>` subagents.

#### Scenario: Everything empty writes nothing

```gherkin
@error
Scenario: Everything empty writes nothing
  Given an opencode project with empty leader, empty default, and all phases empty
  When the user runs "archon init --agent opencode"
  Then no "opencode.json" is created or modified
```

#### Scenario: Non-opencode agent writes nothing

```gherkin
@edge
Scenario: Non-opencode agent writes nothing
  Given a project initialized with "archon init --agent claude"
  When init completes
  Then no "opencode.json" is created or modified
```

#### Scenario: Default set with empty leader still writes subagents

```gherkin
@edge
Scenario: Default set with empty leader still writes subagents
  Given an opencode project with empty leader, "models.default" set, and all phases empty
  When the user runs "archon init --agent opencode"
  Then "opencode.json" is written with all 8 "archon-<phase>" subagents
  And no "agent.archon-leader" entry is written
```

### Requirement: Deterministic, idempotent output

The written `opencode.json` MUST be deterministic and idempotent: map keys sorted,
struct field order fixed, a trailing newline, and an atomic temp-file rename. Running
init or TUI save again with the same configuration MUST yield a byte-identical
`opencode.json`.

#### Scenario: Re-run is byte-identical

```gherkin
@edge
Scenario: Re-run is byte-identical
  Given an opencode project already initialized with per-phase subagents
  When the user runs "archon init --agent opencode" again with the same configuration
  Then the resulting "opencode.json" is byte-identical to the previous run
  And each "archon-<phase>" appears exactly once
```

### Requirement: Preserve existing keys and user agents

The merge MUST preserve pre-existing top-level keys and pre-existing user-defined
agents. Only `archon-leader` and `archon-<phase>` entries MAY be set or overwritten.

#### Scenario: Existing user agent and top-level keys preserved

```gherkin
@edge
Scenario: Existing user agent and top-level keys preserved
  Given an opencode project whose "opencode.json" has unrelated top-level keys and a user-defined agent
  When the user runs "archon init --agent opencode"
  Then every pre-existing top-level key is left unchanged
  And the pre-existing user-defined agent is left unchanged
  And only "archon-leader" and "archon-<phase>" entries are set
```
