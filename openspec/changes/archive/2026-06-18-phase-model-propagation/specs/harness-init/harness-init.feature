Feature: Phase model propagation into the orchestrator template
  Template generation emits a normalized, deterministic phase→model block from
  config.Models on both the init render path and the TUI regeneration path, so the
  orchestrator can request the configured model when delegating each SDD phase.

  Background:
    Given an agent project that uses generated orchestrator instructions

  @happy
  Scenario: Init renders a phase model for a configured phase
    Given a config with "models.phases.propose" set
    When the user runs "archon init --agent claude"
    Then the generated "CLAUDE.md" contains a phase→model block
    And the block lists "propose" with its resolved model

  @happy
  Scenario: TUI regeneration produces the same block as init
    Given a config rendered once by "archon init"
    When the same config is regenerated via the TUI template path
    Then the regenerated orchestrator file contains an identical phase→model block

  @happy
  Scenario: Display name is normalized to an accepted identifier
    Given "models.phases.design" is set to a display string like "Opus 4.8"
    When the orchestrator template is rendered
    Then the "design" line shows a normalized identifier the delegation tool accepts
    And no raw display string appears in the block

  @happy
  Scenario: Phase falls back to the default model
    Given "models.phases.verify" is unset
    And "models.default" is set
    When the template is rendered
    Then the "verify" line shows the normalized default model

  @edge
  Scenario: Phase omitted when no model resolves
    Given "models.phases.apply" is unset
    And "models.default" is unset
    When the template is rendered
    Then the block contains no line for "apply"

  @edge
  Scenario: Multiple configured phases render in canonical order
    Given "models.phases" sets "archive", "explore", and "design"
    When the template is rendered twice
    Then both renders list the phases in canonical SDD order
    And the two renders are byte-identical

  @error
  Scenario: Garbage model value is surfaced
    Given "models.phases.propose" is set to an unresolvable value like "Opues 4.8"
    When the configured models are processed for rendering
    Then the user receives actionable feedback identifying the unknown value
