Feature: archon-judge subagent emission and judge model resolution
  archon init emits a frontmatter-pinned archon-judge subagent (special judgment-day
  wrapper body) and resolves the judge model to mirror verify, while keeping judge a
  valid configurable phase and listing it in CLAUDE.md Phase Models.

  Background:
    Given a claude project initialized with archon

  # Requirement: Per-phase agent file emission
  Scenario: Init writes an agent file per resolvable phase
    Given a claude project with models.default set and models.phases.spec set
    When the user runs "archon init --agent claude"
    Then ".claude/agents/archon-<phase>.md" exists for every phase ResolvePhaseModels returns
    And ".claude/agents/archon-judge.md" is written when judge resolves

  @edge
  Scenario: A phase with no resolvable model is omitted
    Given a claude project with models.default and all models.phases empty except models.phases.spec
    When the user runs "archon init --agent claude"
    Then ".claude/agents/archon-spec.md" is written
    And no agent file is written for any phase ResolvePhaseModels omits

  # Requirement: Functional agent body
  Scenario: Body points the executor at the phase skill
    Given a claude project with a resolvable model for the design phase
    When init writes the claude agents
    Then ".claude/agents/archon-design.md" body references "skills/sdd-design/SKILL.md"
    And the body is non-empty after the frontmatter

  Scenario: Judge body is the judgment-day wrapper
    Given a claude project with a resolvable model for the judge phase
    When init writes the claude agents
    Then ".claude/agents/archon-judge.md" body invokes "judgment-day"
    And the body reports its verdict to "harness-judge"
    And the body does not reference "skills/sdd-judge/SKILL.md"

  # Requirement: Judge model resolution mirrors verify
  Scenario: Judge defaults to the verify model
    Given models.phases.verify resolves to "claude-opus-4-8" and --model-judge is omitted
    When init writes the claude agents
    Then ".claude/agents/archon-judge.md" frontmatter "model" equals "claude-opus-4-8"
    And the model value contains no "/" provider prefix

  @edge
  Scenario: Judge falls back to the default model
    Given models.phases.judge is empty and models.default is provider "anthropic" model "claude-opus-4-8"
    When init writes the claude agents
    Then ".claude/agents/archon-judge.md" frontmatter "model" equals "claude-opus-4-8"

  @edge
  Scenario: Re-init regenerates archon-judge idempotently
    Given a claude project already initialized with archon-judge.md
    When the user runs "archon init --agent claude" again with the same configuration
    Then ".claude/agents/archon-judge.md" is byte-identical to the previous run

  # Requirement: judge is a valid configurable phase
  Scenario: config set then get round-trips the judge model
    Given an initialized claude project
    When the user runs "archon config set models.phases.judge claude-opus-4-8"
    And the user runs "archon config get models.phases.judge"
    Then the reported value equals "claude-opus-4-8"
    And neither command reports an unknown-phase error

  @edge
  Scenario: judge model survives clone and serialization
    Given models.phases.judge set to provider "anthropic" model "claude-opus-4-8"
    When the config is cloned and serialized then reloaded
    Then models.phases.judge resolves to "claude-opus-4-8"

  # Requirement: CLAUDE.md Phase Models lists judge
  Scenario: CLAUDE.md Phase Models includes the judge row
    Given a claude project where judge resolves to "claude-opus-4-8"
    When init writes CLAUDE.md
    Then its "Phase Models" block lists "judge" with model "claude-opus-4-8"
    And the block still names the "archon-<phase>" subagents as the binding
