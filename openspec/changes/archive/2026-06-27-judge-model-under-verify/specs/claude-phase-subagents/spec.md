# Delta for claude-phase-subagents

This delta adds `judge` to the set of phases the claude codegen path emits, so
`.claude/agents/archon-judge.md` becomes a frontmatter-pinned hard gate. Judge is
NOT a generic `sdd-<phase>` executor: it has no `sdd-judge` skill and MUST delegate
the dual review to `judgment-day`, so its emitted body is a special thin wrapper.

## MODIFIED Requirements

### Requirement: Per-phase agent file emission

The writer MUST write one file `.claude/agents/archon-<phase>.md` for each phase
returned by `config.ResolvePhaseModels(cfg.Models)` (chain `Phases[phase]`→`Default`,
omit-when-empty). Because `judge` is now part of `config.PhaseOrder`, a resolvable
`judge` model MUST produce `.claude/agents/archon-judge.md`. A phase with no
resolvable model MUST NOT produce a file.
(Previously: `judge` was excluded from `PhaseOrder`, so `archon-judge.md` MUST NOT be written.)

#### Scenario: Init writes an agent file per resolvable phase

```gherkin
Scenario: Init writes an agent file per resolvable phase
  Given a claude project with models.default set and models.phases.spec set
  When the user runs "archon init --agent claude"
  Then ".claude/agents/archon-<phase>.md" exists for every phase ResolvePhaseModels returns
  And ".claude/agents/archon-judge.md" is written when judge resolves
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

### Requirement: Functional agent body

Each `archon-<phase>.md` body MUST be a functional executor system prompt. For every
phase that HAS an `sdd-<phase>` skill, the body MUST instruct the executor to follow
`skills/sdd-<phase>/SKILL.md` and MUST NOT delegate. For `judge`, which has NO
`sdd-judge` skill, the body MUST instead be a thin wrapper that instructs the executor
to run the `judgment-day` dual review and report the verdict back to `harness-judge`;
the judge body MUST NOT reference a non-existent `skills/sdd-judge/SKILL.md` and MUST
NOT carry the generic "do not delegate; execute yourself" instruction. No body MUST be
an empty stub.
(Previously: every body uniformly referenced `skills/sdd-<phase>/SKILL.md` with a "do not delegate" instruction.)

#### Scenario: Body points the executor at the phase skill

```gherkin
Scenario: Body points the executor at the phase skill
  Given a claude project with a resolvable model for the design phase
  When init writes the claude agents
  Then ".claude/agents/archon-design.md" body references "skills/sdd-design/SKILL.md"
  And the body is non-empty after the frontmatter
```

#### Scenario: Judge body is the judgment-day wrapper

```gherkin
Scenario: Judge body is the judgment-day wrapper
  Given a claude project with a resolvable model for the judge phase
  When init writes the claude agents
  Then ".claude/agents/archon-judge.md" body invokes "judgment-day"
  And the body reports its verdict to "harness-judge"
  And the body does not reference "skills/sdd-judge/SKILL.md"
```

## ADDED Requirements

### Requirement: Judge model resolution mirrors verify

`judge` MUST be a resolvable phase whose default model mirrors `verify`. The
`--model-judge` init flag MUST set `models.phases.judge`; when the flag is omitted it
MUST default to verify's configured value so judge resolves to the same model as
verify. When `models.phases.judge` is unset, `config.ResolvePhaseModels` MUST fall
back to `models.default` (omitting judge only when neither yields a model), identical
to every other phase. The resolved judge model MUST appear in frontmatter as the bare
model id (no `<provider>/` prefix), matching the bare-id rule applied to all
`archon-<phase>.md` files.

#### Scenario: Judge defaults to the verify model

```gherkin
Scenario: Judge defaults to the verify model
  Given models.phases.verify resolves to "claude-opus-4-8" and --model-judge is omitted
  When init writes the claude agents
  Then ".claude/agents/archon-judge.md" frontmatter "model" equals "claude-opus-4-8"
  And the model value contains no "/" provider prefix
```

#### Scenario: Judge falls back to the default model

```gherkin
@edge
Scenario: Judge falls back to the default model
  Given models.phases.judge is empty and models.default is provider "anthropic" model "claude-opus-4-8"
  When init writes the claude agents
  Then ".claude/agents/archon-judge.md" frontmatter "model" equals "claude-opus-4-8"
```

#### Scenario: Re-init regenerates archon-judge idempotently

```gherkin
@edge
Scenario: Re-init regenerates archon-judge idempotently
  Given a claude project already initialized with archon-judge.md
  When the user runs "archon init --agent claude" again with the same configuration
  Then ".claude/agents/archon-judge.md" is byte-identical to the previous run
```

### Requirement: judge is a valid configurable phase

`judge` MUST be a member of `config.ValidPhases` and `config.PhaseOrder`. `archon
config get models.phases.judge` and `archon config set models.phases.judge <model>`
MUST be accepted (no "unknown phase" error), and the value MUST round-trip through
config serialization and `Config.Clone()`. Adding `judge` MUST NOT alter the SDD phase
ORDER or the `harness-workflow` state-machine sequence.

#### Scenario: config set then get round-trips the judge model

```gherkin
Scenario: config set then get round-trips the judge model
  Given an initialized claude project
  When the user runs "archon config set models.phases.judge claude-opus-4-8"
  And the user runs "archon config get models.phases.judge"
  Then the reported value equals "claude-opus-4-8"
  And neither command reports an unknown-phase error
```

#### Scenario: judge model survives clone and serialization

```gherkin
@edge
Scenario: judge model survives clone and serialization
  Given models.phases.judge set to provider "anthropic" model "claude-opus-4-8"
  When the config is cloned and serialized then reloaded
  Then models.phases.judge resolves to "claude-opus-4-8"
```

### Requirement: CLAUDE.md Phase Models lists judge

The generated `CLAUDE.md` "## Phase Models" block, derived from `config.PhaseOrder`,
MUST include a `judge` line showing the resolved judge model (e.g.
`judge → claude-opus-4-8` when judge mirrors verify). The block MUST keep naming the
`archon-<phase>` subagents as the binding hard gate.

#### Scenario: CLAUDE.md Phase Models includes the judge row

```gherkin
Scenario: CLAUDE.md Phase Models includes the judge row
  Given a claude project where judge resolves to "claude-opus-4-8"
  When init writes CLAUDE.md
  Then its "Phase Models" block lists "judge" with model "claude-opus-4-8"
  And the block still names the "archon-<phase>" subagents as the binding
```
