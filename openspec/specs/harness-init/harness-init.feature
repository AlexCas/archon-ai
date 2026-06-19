Feature: Harness initialization UX
  archon init bootstraps any project, protects existing orchestrator files,
  and configures web testing and models.

  Scenario: Blank project initialized with an explicit agent
    Given a project directory with no agent folder
    When the user runs "archon init --agent claude"
    Then the ".claude" directory is created
    And ".archon/config.yaml" is created
    And "CLAUDE.md" is created

  Scenario: User declines to replace an existing orchestrator file
    Given a project with an existing "CLAUDE.md"
    When the user runs "archon init --agent claude" and answers "n"
    Then the existing "CLAUDE.md" is left unchanged
    And no ".archon" directory is created

  Scenario: Force replaces without prompting
    Given a project with an existing "CLAUDE.md"
    When the user runs "archon init --agent claude --force"
    Then "CLAUDE.md" is regenerated
    And init completes successfully

  Scenario: Enabling Playwright at init
    Given a web project
    When the user runs "archon init --agent claude --playwright"
    Then "playwright.enabled" is true in ".archon/config.yaml"

  @ux
  Scenario: Selecting a static model in the TUI
    Given the Models tab is focused on the default model input
    When the user cycles the static catalog with "ctrl+n"
    Then the input is set to a Claude or Opencode Go model from the catalog

  @ux
  Scenario: Typing a free-form model
    Given the Models tab default model input is empty
    When the user types "some-custom-model"
    Then the value is accepted
    And a non-blocking warning may be shown

  @happy
  Scenario: Init records real frontmatter versions
    Given embedded skills whose "SKILL.md" frontmatter declares versions like "2.0" and "3.0"
    When the user runs "archon init --agent claude"
    Then each "skill_inventory" entry records that skill's real frontmatter version
    And no inventory entry uses a hardcoded version

  @edge
  Scenario: Missing frontmatter version is handled
    Given an embedded skill whose "SKILL.md" declares no metadata.version
    When the user runs "archon init --agent claude"
    Then that skill is still recorded in "skill_inventory"
    And init does not abort

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

  # --- opencode-leader-mode ---

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
