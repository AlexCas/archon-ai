Feature: TUI Tab Ordering and Default Landing Tab
  The archon TUI MUST open on the Agent tab and render tabs in the order
  Agent, Models, Judge, Mutation Testing, Playwright so that the user is
  guided to configure the agent before the models.

  Background:
    Given the archon TUI is initialized with a default config

  @happy
  Scenario: TUI opens on the Agent tab
    When the model is created via NewModel
    Then the active tab is AgentTab

  @happy
  Scenario: Header renders tabs in correct order
    When the view is rendered
    Then the tab strip shows "Agent" as the first tab
    And the tab strip shows "Models" as the second tab
    And the tab strip shows "Judge" as the third tab
    And the tab strip shows "Mutation Testing" as the fourth tab
    And the tab strip shows "Playwright" as the fifth tab

  @happy
  Scenario: Tab key advances from Agent to Models
    Given the archon TUI is initialized with AgentTab active
    When the user presses the Tab key
    Then the active tab is ModelsTab

  @edge
  Scenario: Shift+Tab wraps from Agent to Playwright
    Given the archon TUI is initialized with AgentTab active
    When the user presses the Shift+Tab key
    Then the active tab is PlaywrightTab

  @edge
  Scenario: Full tab cycle returns to Agent
    Given the archon TUI is initialized with AgentTab active
    When the user presses Tab tabCount*3 times
    Then the active tab is AgentTab
