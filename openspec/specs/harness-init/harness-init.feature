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

  @happy
  Scenario: Installed opencode shows the live catalog
    Given the "opencode" CLI is installed on PATH
    When the user opens the Models view
    Then the offered opencode models are enumerated live from the opencode CLI
    And the stale curated opencode list is not shown

  @happy
  Scenario: Only installed agents' models are offered
    Given the "opencode" CLI is not installed on PATH
    When the user cycles the catalog in the Models view
    Then no opencode models appear in the offered catalog
    And models for installed agents remain offered

  @edge
  Scenario: Detection is cached once per Models view
    Given the Models view has been opened and detection has run once
    When the user cycles models and types repeatedly
    Then detection does not run again for that session
    And it never runs during "archon init"

  @error
  Scenario: Live enumeration error falls back silently
    Given the "opencode" CLI is installed but enumeration fails, times out, or returns unparseable output
    When the user opens the Models view
    Then the curated opencode list is offered as a silent fallback
    And the TUI is neither blocked nor shown an error

  @happy
  Scenario: Free-form entry and advisory behavior unchanged
    Given the Models view default model input is empty
    When the user types "some-custom-model"
    Then the value is accepted and a non-blocking warning may be shown
    And NormalizeModel and Validate behave exactly as before this feature

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
