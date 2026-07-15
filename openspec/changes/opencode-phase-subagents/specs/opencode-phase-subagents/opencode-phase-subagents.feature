Feature: Opencode per-phase subagents
  For an opencode project, archon init and the TUI save path write one subagent per SDD
  phase (keyed archon-<phase>) into opencode.json, each carrying its resolved per-phase
  model written verbatim, with deterministic, idempotent output that preserves existing
  keys and user-defined agents.

  Background:
    Given an opencode project that uses generated "AGENTS.md" orchestrator instructions

  @happy
  Scenario: Init writes 8 per-phase subagents
    Given an opencode project with per-phase models configured for every phase
    When the user runs "archon init --agent opencode"
    Then "opencode.json" contains an "agent.archon-<phase>" for each of the 8 PhaseOrder phases
    And each subagent's "model" equals the configured per-phase model verbatim
    And no "agent.archon-judge" entry is written

  @happy
  Scenario: Phase with its own model uses it verbatim
    Given "models.phases.spec" set to "anthropic/claude-opus-4-20250514"
    When init writes the opencode agents
    Then "agent.archon-spec.model" equals "anthropic/claude-opus-4-20250514" verbatim

  @edge
  Scenario: Phase without a model falls back to default
    Given "models.phases.tasks" is empty and "models.default" is "anthropic/claude-sonnet-4-20250514"
    When init writes the opencode agents
    Then "agent.archon-tasks" is still present
    And its "model" equals "models.default" verbatim

  @edge
  Scenario: Phase and default empty fall back to leader
    Given "models.phases.verify" and "models.default" are both empty
    And "models.leader" is "anthropic/claude-opus-4-20250514"
    When init writes the opencode agents
    Then "agent.archon-verify" is still present
    And its "model" equals "models.leader" verbatim

  @happy
  Scenario: Subagent carries the fixed fields
    Given an opencode project with any non-empty model configuration
    When init writes the opencode agents
    Then every "agent.archon-<phase>" has "mode" equal to "subagent"
    And every "agent.archon-<phase>" has "hidden" equal to true
    And every "agent.archon-<phase>" has a "model", a "description", and a "prompt"

  @error
  Scenario: Everything empty writes nothing
    Given an opencode project with empty leader, empty default, and all phases empty
    When the user runs "archon init --agent opencode"
    Then no "opencode.json" is created or modified

  @edge
  Scenario: Non-opencode agent writes nothing
    Given a project initialized with "archon init --agent claude"
    When init completes
    Then no "opencode.json" is created or modified

  @edge
  Scenario: Default set with empty leader still writes subagents
    Given an opencode project with empty leader, "models.default" set, and all phases empty
    When the user runs "archon init --agent opencode"
    Then "opencode.json" is written with all 8 "archon-<phase>" subagents
    And no "agent.archon-leader" entry is written

  @edge
  Scenario: Re-run is byte-identical
    Given an opencode project already initialized with per-phase subagents
    When the user runs "archon init --agent opencode" again with the same configuration
    Then the resulting "opencode.json" is byte-identical to the previous run
    And each "archon-<phase>" appears exactly once

  @edge
  Scenario: Existing user agent and top-level keys preserved
    Given an opencode project whose "opencode.json" has unrelated top-level keys and a user-defined agent
    When the user runs "archon init --agent opencode"
    Then every pre-existing top-level key is left unchanged
    And the pre-existing user-defined agent is left unchanged
    And only "archon-leader" and "archon-<phase>" entries are set
