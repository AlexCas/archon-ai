Feature: Dynamic model selection with free-form fallback
  The Models view offers a catalog that reflects the agent CLIs installed on PATH,
  enumerates opencode models live when the opencode CLI is present, and falls back
  silently to curated lists on any failure — while free-form entry and the advisory
  Validate/NormalizeModel behavior remain unchanged.

  Background:
    Given the archon TUI Models view

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
