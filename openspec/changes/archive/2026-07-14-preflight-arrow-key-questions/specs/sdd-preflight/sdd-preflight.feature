Feature: SDD Preflight as Per-Group Arrow-Key Questions
  The ARCHON orchestrator MUST ask the SDD session preflight as five separate
  arrow-key questions (one per group A–E) instead of a single Spanish code block,
  so that users select options with arrow keys rather than typing error-prone codes.

  Background:
    Given an archon-initialized CLAUDE.md that contains the preflight instructions

  @happy
  Scenario: Five groups asked as separate questions
    Given the orchestrator starts an SDD session with no prior preflight decision
    When the orchestrator issues the preflight interaction
    Then the orchestrator asks group A (Ritmo) as an arrow-key question
    And the orchestrator asks group B (Artefactos) as an arrow-key question
    And the orchestrator asks group C (PRs) as an arrow-key question
    And the orchestrator asks group D (Revisión) as an arrow-key question
    And the orchestrator asks group E (Playwright) as an arrow-key question

  @happy
  Scenario: Recommended option is pre-selected default
    Given the orchestrator renders each preflight group question
    When the user accepts all defaults without changing any selection
    Then group A defaults to "Interactivo"
    And group B defaults to "OpenSpec"
    And group C defaults to "Preguntarme"
    And group D defaults to "400 líneas"
    And group E defaults to "No"

  @edge
  Scenario: D "Otro" triggers free-text follow-up
    Given the orchestrator is asking group D (Revisión)
    When the user selects "Otro"
    Then the orchestrator asks a free-text question for the custom line count
    And the orchestrator caches the entered number as the review budget

  @happy
  Scenario: Group E asked for a non-web project
    Given an archon-initialized CLAUDE.md for a non-web Go project
    When the orchestrator starts an SDD session
    Then the orchestrator asks group E (Playwright) as an arrow-key question

  @edge
  Scenario: Group E asked for a web project
    Given an archon-initialized CLAUDE.md for a web project
    When the orchestrator starts an SDD session
    Then the orchestrator asks group E (Playwright) as an arrow-key question

  @error
  Scenario: No legacy code block in generated CLAUDE.md
    Given the rendered CLAUDE.md preflight section
    When the content is inspected for legacy artifacts
    Then the text does not contain "Antes de continuar con SDD"
    And the text does not contain answer codes like "A1" or "B1"
    And the text does not contain a "usar recomendado" fast-path instruction

  @error
  Scenario: Hard gate blocks SDD phase when no preflight decision exists
    Given the orchestrator has no cached preflight decisions for the session
    When the orchestrator receives any SDD command
    Then the orchestrator STOPS and asks the five preflight questions before proceeding

  @happy
  Scenario: Preflight choices cached and echoed into later phase prompts
    Given the user has answered all five preflight questions
    When the orchestrator transitions to any subsequent SDD phase
    Then the orchestrator echoes the cached preflight choices into the phase prompt
    And the choices are recorded in SESSION_STATUS.md
