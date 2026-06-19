Feature: Opencode archon-leader mode and configurable leader model
  archon init writes an additive, idempotent "archon-leader" primary agent into the
  project opencode.json using a configurable leader model, with the TUI save path
  producing the same merge and archon update staying skill-only.

  Background:
    Given an opencode project that uses generated "AGENTS.md" orchestrator instructions

  @happy
  Scenario: Leader model survives clone and round-trip
    Given "models.leader" set to "anthropic/claude-sonnet-4-20250514"
    When the config is cloned and serialized then reloaded
    Then "models.leader" equals "anthropic/claude-sonnet-4-20250514" verbatim

  @happy
  Scenario: Init writes the archon-leader agent
    Given an opencode project with "models.leader" set and no "opencode.json"
    When the user runs "archon init --agent opencode"
    Then "opencode.json" sets "agent.archon-leader" with mode "primary"
    And its "prompt" is "{file:./AGENTS.md}" and "model" equals "models.leader"
    And the "opencode.json" path is registered for rollback

  @edge
  Scenario: Merge into an existing opencode.json preserves other keys
    Given an opencode project whose "opencode.json" has unrelated keys and agents
    When the user runs "archon init --agent opencode"
    Then "agent.archon-leader" is added
    And every pre-existing key and agent is left unchanged
    And no "default_agent" key is written

  @edge
  Scenario: Re-running init is idempotent
    Given an opencode project already initialized with "archon-leader"
    When the user runs "archon init --agent opencode" again
    Then "agent.archon-leader" appears exactly once with the same content
    And no other key drifts

  @edge
  Scenario: Non-opencode agent writes no opencode.json
    Given a project initialized with "archon init --agent claude"
    When init completes
    Then no "opencode.json" is created or modified

  @error
  Scenario: Empty leader model writes nothing
    Given an opencode project with an empty "models.leader"
    When the user runs "archon init --agent opencode"
    Then no "opencode.json" is created
    And no archon-leader agent is written

  @happy
  Scenario: TUI save matches init merge result
    Given an opencode project and a chosen leader model
    When the leader-model field is saved via the TUI Models tab
    Then the resulting "agent.archon-leader" equals what "archon init" would produce

  @edge
  Scenario: Update leaves the opencode agent untouched
    Given an opencode project with an existing "agent.archon-leader"
    When the user runs "archon update"
    Then "opencode.json" is not written or rewritten
