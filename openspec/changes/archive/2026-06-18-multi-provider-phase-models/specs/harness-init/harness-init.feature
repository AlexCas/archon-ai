Feature: Multi-provider per-phase model normalization
  Model normalization recognizes curated identifiers from Claude, Gemini, OpenAI, and
  Opencode, with a fixed cross-provider precedence, so the phase→model block renders for
  any supported provider and stays advisory toward unknown values.

  Background:
    Given an agent project that uses generated orchestrator instructions

  @happy
  Scenario: Display name is normalized to an accepted identifier
    Given "models.phases.design" is set to a display string like "Opus 4.8"
    When the orchestrator template is rendered
    Then the "design" line shows a normalized identifier the delegation tool accepts
    And no raw display string appears in the block

  @happy
  Scenario: Gemini model normalizes to its catalog id
    Given "models.phases.spec" is set to a curated Gemini catalog id
    When the orchestrator template is rendered
    Then the "spec" line shows that Gemini catalog id as-is

  @happy
  Scenario: OpenAI model normalizes to its catalog id
    Given "models.phases.tasks" is set to a curated OpenAI catalog id
    When the orchestrator template is rendered
    Then the "tasks" line shows that OpenAI catalog id as-is

  @happy
  Scenario: Opencode model normalizes to its catalog id
    Given "models.phases.apply" is set to a curated Opencode catalog id
    When the orchestrator template is rendered
    Then the "apply" line shows that Opencode catalog id as-is

  @edge
  Scenario: Whole-token guard rejects a containing substring
    Given "models.phases.verify" is set to "octopus"
    When the value is normalized
    Then it does not match the Claude "opus" family

  @edge
  Scenario: Colliding value resolves by fixed precedence
    Given a value that matches both Claude and a later provider
    When the value is normalized
    Then it resolves to the Claude canonical form

  @happy
  Scenario: Non-Claude default renders an identical block across paths
    Given "models.default" is set to a curated non-Claude catalog id
    When the file is rendered via "archon init" and via the TUI regenerate path
    Then both produce a non-empty "## Phase Models" block
    And the two blocks are byte-identical

  @error
  Scenario: Unresolvable typo is omitted but not rejected
    Given "models.phases.propose" is set to an unresolvable value like "Opues 4.8"
    When the configured models are processed for rendering
    Then "propose" is omitted from the block
    And the value is accepted with an advisory warning, not rejected
