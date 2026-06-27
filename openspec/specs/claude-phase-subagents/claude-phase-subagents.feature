Feature: Claude per-phase subagent hard gate
  For a claude project, archon init and the TUI save path write one
  .claude/agents/archon-<phase>.md per resolvable SDD phase, binding its
  resolved FullID via frontmatter model: so per-phase model selection is a
  hard gate rather than advice.

  @happy
  Scenario: Init writes an agent file per resolvable phase
    Given a claude project with models.default set and models.phases.spec set
    When the user runs "archon init --agent claude"
    Then ".claude/agents/archon-<phase>.md" exists for every phase ResolvePhaseModels returns
    And no ".claude/agents/archon-judge.md" file is written

  @edge
  Scenario: A phase with no resolvable model is omitted
    Given a claude project with models.default and all models.phases empty except models.phases.spec
    When the user runs "archon init --agent claude"
    Then ".claude/agents/archon-spec.md" is written
    And no agent file is written for any phase ResolvePhaseModels omits

  @happy
  Scenario: Frontmatter model strips the provider prefix
    Given models.phases.spec is provider "anthropic" model "claude-opus-4-8"
    When init writes the claude agents
    Then ".claude/agents/archon-spec.md" frontmatter "model" equals "claude-opus-4-8"
    And the model value contains no "/" provider prefix

  @edge
  Scenario: Phase falls back to the default model
    Given models.phases.tasks is empty and models.default is provider "anthropic" model "claude-sonnet-4-6"
    When init writes the claude agents
    Then ".claude/agents/archon-tasks.md" frontmatter "model" equals "claude-sonnet-4-6"

  @edge
  Scenario: Bare alias is passed through unchanged
    Given models.phases.spec is the bare alias "opus" with no provider
    When init writes the claude agents
    Then ".claude/agents/archon-spec.md" frontmatter "model" equals "opus"

  @happy
  Scenario: Body points the executor at the phase skill
    Given a claude project with a resolvable model for the design phase
    When init writes the claude agents
    Then ".claude/agents/archon-design.md" body references "skills/sdd-design/SKILL.md"
    And the body is non-empty after the frontmatter

  @error
  Scenario: Nothing resolvable writes nothing
    Given a claude project with empty default and all phases empty
    When the user runs "archon init --agent claude"
    Then no ".claude/agents/archon-*.md" file is created

  @edge
  Scenario: Non-claude agent writes no claude agent files
    Given an opencode project with resolvable phase models
    When the user runs "archon init --agent opencode"
    Then no ".claude/agents/archon-*.md" file is created

  @edge
  Scenario: Re-run is byte-identical and preserves user files
    Given a .claude/agents directory with an unrelated user file and archon agents already written
    When init runs again with the same configuration
    Then each archon-<phase>.md is byte-identical to the previous run
    And the unrelated user file is unchanged

  @happy
  Scenario: Undo removes the generated agent files
    Given the user ran "archon init --agent claude" and agent files were written
    When the user runs "archon undo"
    Then every generated ".claude/agents/archon-<phase>.md" is removed

  @happy
  Scenario: CLAUDE.md names subagents as the hard gate
    Given a claude project with resolvable phase models
    When init writes CLAUDE.md
    Then its "Phase Models" block names the "archon-<phase>" subagents as the binding
    And the block does not call model selection "advisory"

  @edge
  Scenario: AGENTS.md points at opencode.json
    Given an opencode project with resolvable phase models
    When init writes AGENTS.md
    Then its "Phase Models" block states the binding lives in "opencode.json"

  @happy
  Scenario: CLAUDE.md routes delegation to the named subagent
    Given a claude project with resolvable phase models
    When init writes CLAUDE.md
    Then its delegation rule names the "archon-<phase>" subagent as the delegation target per phase
    And it instructs the leader not to pass a per-call model parameter

  @edge
  Scenario: AGENTS.md routes delegation to the named subagent
    Given an opencode project with resolvable phase models
    When init writes AGENTS.md
    Then its delegation rule names the "archon-<phase>" subagent as the delegation target per phase
