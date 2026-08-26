Feature: Graphify Integration — TUI Tab (Delta for graphify-tui-tab)
  Adds an interactive Graphify tab to archon tui, exposing all five
  config.Graphify fields. Modifies the R-05 group G mapping paragraph to
  reference the new tab. All other graphify-integration requirements are
  unchanged.

  Background:
    Given an archon project with a valid .archon/config.yaml

  # ── R-19: TUI Tab ─────────────────────────────────────────────────────────────

  @happy
  Scenario: Graphify tab appears in archon tui tab bar
    Given archon tui is running
    When the user navigates through the tab bar
    Then a "Graphify" tab is listed after the "Impeccable" tab

  @happy
  Scenario: All five config.Graphify fields are visible in the Graphify tab
    Given the user has selected the Graphify tab
    When the tab renders
    Then the tab displays controls for enabled, auto_install, semantic, version, and output_dir

  @happy
  Scenario: Toggling enabled with Enter updates the field
    Given the Graphify tab is focused at index 0 (enabled)
    When the user presses Enter
    Then the enabled toggle flips to its opposite value

  @happy
  Scenario: Toggling auto_install with Space updates the field
    Given the Graphify tab is focused at index 1 (auto_install)
    When the user presses Space
    Then the auto_install toggle flips to its opposite value

  @happy
  Scenario: Toggling semantic at focus index 2 updates the field
    Given the Graphify tab is focused at index 2 (semantic)
    When the user presses Enter
    Then the semantic toggle flips to its opposite value

  @happy
  Scenario: Editing version text input updates config on save
    Given the Graphify tab is focused at index 3 (version)
    When the user types "v0.9.50" and saves
    Then config.Graphify.Version equals "v0.9.50"

  @happy
  Scenario: Editing output_dir text input updates config on save
    Given the Graphify tab is focused at index 4 (output_dir)
    When the user types ".archon/custom" and saves
    Then config.Graphify.OutputDir equals ".archon/custom"

  @edge
  Scenario: Blank version input coerces to DefaultGraphifyVersion on save
    Given the Graphify tab has an empty version field
    When the user saves
    Then config.Graphify.Version equals "v0.9.45"
    And the saved YAML does not contain an empty version key

  @edge
  Scenario: Blank output_dir input coerces to DefaultGraphifyOutputDir on save
    Given the Graphify tab has an empty output_dir field
    When the user saves
    Then config.Graphify.OutputDir equals ".archon/graphify"
    And the saved YAML does not contain an empty output_dir key

  @happy
  Scenario: Graphify tab wired in model — Shift+Tab from AgentTab wraps to GraphifyTab
    Given archon tui is running with GraphifyTab as the last tab
    When the user presses Shift+Tab from AgentTab
    Then the focused tab is GraphifyTab

  @happy
  Scenario: renderTabs lists all tab labels in order including Graphify
    Given archon tui renders the tab bar
    When renderTabs is called
    Then the label list ends with "Impeccable" followed by "Graphify"

  @happy
  Scenario: Tab resize propagates to Graphify tab inputs
    Given the Graphify tab is active
    When a WindowSizeMsg is received with a new width
    Then the Graphify tab's text inputs resize to the new width

  # ── R-05 (MODIFIED): Preflight Group G mapping paragraph ─────────────────────

  @happy
  Scenario: Group G mapping paragraph references both the init flag and the TUI tab
    Given internal/initcmd/templates.go and root CLAUDE.md are current
    When the orchestrator renders CLAUDE.md
    Then the group G mapping paragraph contains "The `--graphify` flag at init time or the Graphify tab in `archon tui` set the same value."
    And the file still contains "¿Activar Graphify para análisis de grafo de código?"
    And the group count reads "seven" and the range reads "A-G"
