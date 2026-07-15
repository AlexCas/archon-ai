# claude-phase-subagents Specification

## Purpose

For a **claude** project, `archon init` and the TUI save path MUST write one
`.claude/agents/archon-<phase>.md` file per resolvable SDD phase, each carrying its
resolved per-phase model in YAML frontmatter `model: <FullID>` and a functional body
that follows `skills/sdd-<phase>/SKILL.md`. Resolution reuses
`config.ResolvePhaseModels`, making these subagents the binding hard gate that the
generated `CLAUDE.md` "Phase Models" block names. This mirrors `opencode-phase-writers`.

## Requirements

### Requirement: Per-phase agent file emission

The writer MUST write one file `.claude/agents/archon-<phase>.md` for each phase returned
by `config.ResolvePhaseModels(cfg.Models)` (chain `Phases[phase]`→`Default`,
omit-when-empty). Because `judge` is never in `PhaseOrder`, `archon-judge.md` MUST NOT be
written. A phase with no resolvable model MUST NOT produce a file.

#### Scenario: Init writes an agent file per resolvable phase

```gherkin
Scenario: Init writes an agent file per resolvable phase
  Given a claude project with models.default set and models.phases.spec set
  When the user runs "archon init --agent claude"
  Then ".claude/agents/archon-<phase>.md" exists for every phase ResolvePhaseModels returns
  And no ".claude/agents/archon-judge.md" file is written
```

#### Scenario: A phase with no resolvable model is omitted

```gherkin
@edge
Scenario: A phase with no resolvable model is omitted
  Given a claude project with models.default and all models.phases empty except models.phases.spec
  When the user runs "archon init --agent claude"
  Then ".claude/agents/archon-spec.md" is written
  And no agent file is written for any phase ResolvePhaseModels omits
```

### Requirement: Frontmatter model is the resolved FullID

Each `archon-<phase>.md` file MUST carry YAML frontmatter whose `model` field equals the
`PhaseModel.Model` (FullID) that `ResolvePhaseModels` produces for that phase — identical
to the value the `CLAUDE.md` "Phase Models" section names for that phase.

#### Scenario: Frontmatter model matches the resolved FullID

```gherkin
Scenario: Frontmatter model matches the resolved FullID
  Given models.phases.spec is provider "anthropic" model "claude-opus-4-8"
  When init writes the claude agents
  Then ".claude/agents/archon-spec.md" frontmatter "model" equals "anthropic/claude-opus-4-8"
```

#### Scenario: Phase falls back to the default model

```gherkin
@edge
Scenario: Phase falls back to the default model
  Given models.phases.tasks is empty and models.default is provider "anthropic" model "claude-sonnet-4-6"
  When init writes the claude agents
  Then ".claude/agents/archon-tasks.md" frontmatter "model" equals "anthropic/claude-sonnet-4-6"
```

### Requirement: Functional agent body

Each `archon-<phase>.md` body MUST be a functional executor system prompt that instructs
the executor to follow `skills/sdd-<phase>/SKILL.md` for its phase. The body MUST NOT be
an empty stub.

#### Scenario: Body points the executor at the phase skill

```gherkin
Scenario: Body points the executor at the phase skill
  Given a claude project with a resolvable model for the design phase
  When init writes the claude agents
  Then ".claude/agents/archon-design.md" body references "skills/sdd-design/SKILL.md"
  And the body is non-empty after the frontmatter
```

### Requirement: No-op condition

No `.claude/agents/archon-*.md` file MUST be written when `ResolvePhaseModels` returns no
phases (no resolvable models) OR when the agent is not claude.

#### Scenario: Nothing resolvable writes nothing

```gherkin
@error
Scenario: Nothing resolvable writes nothing
  Given a claude project with empty default and all phases empty
  When the user runs "archon init --agent claude"
  Then no ".claude/agents/archon-*.md" file is created
```

#### Scenario: Non-claude agent writes no claude agent files

```gherkin
@edge
Scenario: Non-claude agent writes no claude agent files
  Given an opencode project with resolvable phase models
  When the user runs "archon init --agent opencode"
  Then no ".claude/agents/archon-*.md" file is created
```

### Requirement: Deterministic output and preservation

The written files MUST be deterministic and idempotent (fixed frontmatter field order,
trailing newline, atomic temp+rename) so re-runs with the same configuration are
byte-identical. The writer MUST manage only `archon-<phase>.md` files and MUST NOT modify
or delete unrelated user-defined files under `.claude/agents/`.

#### Scenario: Re-run is byte-identical and preserves user files

```gherkin
@edge
Scenario: Re-run is byte-identical and preserves user files
  Given a .claude/agents directory with an unrelated user file and archon agents already written
  When init runs again with the same configuration
  Then each archon-<phase>.md is byte-identical to the previous run
  And the unrelated user file is unchanged
```

### Requirement: Generated paths are rollback-registered

Every `.claude/agents/archon-<phase>.md` path the writer produces MUST be registered in the
rollback manifest so `archon undo` removes the generated agent files.

#### Scenario: Undo removes the generated agent files

```gherkin
Scenario: Undo removes the generated agent files
  Given the user ran "archon init --agent claude" and agent files were written
  When the user runs "archon undo"
  Then every generated ".claude/agents/archon-<phase>.md" is removed
```

### Requirement: Orchestrator doc names the binding mechanism

The generated `CLAUDE.md` "## Phase Models" block MUST name the `archon-<phase>` subagents
as the binding mechanism (hard gate) and MUST NOT describe model selection as "advisory".
The generated `AGENTS.md` (opencode) variant MUST state that the binding lives in
`opencode.json`.

#### Scenario: CLAUDE.md names subagents as the hard gate

```gherkin
Scenario: CLAUDE.md names subagents as the hard gate
  Given a claude project with resolvable phase models
  When init writes CLAUDE.md
  Then its "Phase Models" block names the "archon-<phase>" subagents as the binding
  And the block does not call model selection "advisory"
```

#### Scenario: AGENTS.md points at opencode.json

```gherkin
@edge
Scenario: AGENTS.md points at opencode.json
  Given an opencode project with resolvable phase models
  When init writes AGENTS.md
  Then its "Phase Models" block states the binding lives in "opencode.json"
```

### Requirement: Orchestrator delegates phases by named subagent

The generated orchestrator doc MUST instruct the leader to delegate each phase to the
named `archon-<phase>` subagent (whose own model is the hard gate), NOT to a generic
agent. The instruction MUST state that the leader MUST NOT pass a per-call model
parameter (so the subagent's bound model wins). This applies to both `CLAUDE.md` and
`AGENTS.md`.

#### Scenario: CLAUDE.md routes delegation to the named subagent

```gherkin
Scenario: CLAUDE.md routes delegation to the named subagent
  Given a claude project with resolvable phase models
  When init writes CLAUDE.md
  Then its delegation rule names the "archon-<phase>" subagent as the delegation target per phase
  And it instructs the leader not to pass a per-call model parameter
```

#### Scenario: AGENTS.md routes delegation to the named subagent

```gherkin
@edge
Scenario: AGENTS.md routes delegation to the named subagent
  Given an opencode project with resolvable phase models
  When init writes AGENTS.md
  Then its delegation rule names the "archon-<phase>" subagent as the delegation target per phase
```
