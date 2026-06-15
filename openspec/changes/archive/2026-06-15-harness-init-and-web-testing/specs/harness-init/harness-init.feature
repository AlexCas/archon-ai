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
