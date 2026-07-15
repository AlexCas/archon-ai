# Delta for TUI Tabs

## Purpose

This is a behavioral spec for the TUI tab ordering and default landing tab. No
`openspec/specs/` capability governs tab order (it is UI behavior internal to
`internal/tui`). This delta captures the acceptance criteria as a full behavioral
spec for the change.

## ADDED Requirements

### Requirement: Default Landing Tab

The TUI MUST open with the Agent tab active when launched via `archon tui`.

The `NewModel` constructor SHALL set `activeTab` to `AgentTab` so that the first
screen the user sees is the Agent configuration tab.

#### Scenario: TUI opens on the Agent tab

```gherkin
Scenario: TUI opens on the Agent tab
  Given the archon TUI is initialized with a default config
  When the model is created via NewModel
  Then the active tab is AgentTab
```

### Requirement: Tab Header Order

The TUI header MUST render tabs in the order: Agent, Models, Judge, Mutation Testing,
Playwright (left to right).

The `Tab` enum SHALL assign `AgentTab = 0`, `ModelsTab = 1`, `JudgeTab = 2`,
`MutationTab = 3`, `PlaywrightTab = 4`. The header label slice SHALL match this order
so that the rendered tab strip reads `Agent | Models | Judge | Mutation Testing | Playwright`.

#### Scenario: Header renders tabs in correct order

```gherkin
Scenario: Header renders tabs in correct order
  Given the archon TUI is initialized with a default config
  When the view is rendered
  Then the tab strip shows "Agent" as the first tab
  And the tab strip shows "Models" as the second tab
  And the tab strip shows "Judge" as the third tab
  And the tab strip shows "Mutation Testing" as the fourth tab
  And the tab strip shows "Playwright" as the fifth tab
```

### Requirement: Tab Key Navigation from Agent

The Tab key MUST advance from Agent (the default first tab) to Models (the second tab).

From `AgentTab`, one Tab keypress SHALL set `activeTab` to `ModelsTab`.

#### Scenario: Tab key advances from Agent to Models

```gherkin
Scenario: Tab key advances from Agent to Models
  Given the archon TUI is initialized with AgentTab active
  When the user presses the Tab key
  Then the active tab is ModelsTab
```

### Requirement: Shift+Tab Navigation Wraps to Playwright

Shift+Tab from Agent (the first tab) MUST wrap around to Playwright (the last tab).

From `AgentTab`, one Shift+Tab keypress SHALL set `activeTab` to `PlaywrightTab`.

#### Scenario: Shift+Tab wraps from Agent to Playwright

```gherkin
Scenario: Shift+Tab wraps from Agent to Playwright
  Given the archon TUI is initialized with AgentTab active
  When the user presses the Shift+Tab key
  Then the active tab is PlaywrightTab
```

### Requirement: Full-Cycle Navigation Returns to Default Tab

Cycling through all tabs an exact whole number of times MUST return to the active tab
at the start of the cycle.

After `tabCount * N` Tab keypresses (N >= 1), the `activeTab` SHALL equal the tab that
was active before the cycle began.

#### Scenario: Full tab cycle returns to Agent

```gherkin
Scenario: Full tab cycle returns to Agent
  Given the archon TUI is initialized with AgentTab active
  When the user presses Tab tabCount*3 times
  Then the active tab is AgentTab
```
